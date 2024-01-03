package api

import (
	"encoding/base64"
	"fmt"
	"github.com/ppoonk/AirGo/global"
	"github.com/ppoonk/AirGo/model"
	"github.com/ppoonk/AirGo/service"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AGGetNodeInfo(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	id := ctx.Query("id")
	//node, _, err := service.CommonSqlFind[model.Node, string, model.AGNodeInfo](fmt.Sprintf("id = %s", id))

	var node model.Node
	err := global.DB.Where("id = ?", id).Preload("Access").First(&node).Error

	if err != nil {
		global.Logrus.Error("AGGetNodeInfo error,id="+id, err.Error())
		return
	}
	//处理探针
	global.GoroutinePool.Submit(func() {
		//取消离线节点的通知状态
		global.LocalCache.Delete(fmt.Sprintf("%d%s", node.ID, global.NodeStatusIsNotified))

		cacheStatus, ok := global.LocalCache.Get(id + global.NodeStatus)
		if ok && cacheStatus != nil {
			oldStatus := cacheStatus.(model.NodeStatus)
			oldStatus.Status = true
			global.LocalCache.Set(id+global.NodeStatus, oldStatus, 2*time.Minute) //2分钟后过期

		} else {
			var status model.NodeStatus
			status.Status = true
			status.ID, _ = strconv.ParseInt(id, 64, 10)
			global.LocalCache.Set(id+global.NodeStatus, status, 2*time.Minute) //2分钟后过期
		}
	})
	//处理ss节点加密
	if node.NodeType == "shadowsocks" {
		switch node.Scy {
		case "2022-blake3-aes-128-gcm":
			node.ServerKey = base64.StdEncoding.EncodeToString([]byte(node.ServerKey[:16]))
		default:
			node.ServerKey = base64.StdEncoding.EncodeToString([]byte(node.ServerKey))
		}
	}
	//etag
	EtagHandler(node, ctx)
}

func AGReportNodeStatus(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	var AGNodeStatus model.AGNodeStatus
	err := ctx.ShouldBind(&AGNodeStatus)
	if err != nil {
		global.Logrus.Error("error", err.Error())
		return
	}
	//处理探针
	global.GoroutinePool.Submit(func() {
		cacheStatus, ok := global.LocalCache.Get(strconv.FormatInt(AGNodeStatus.ID, 10) + global.NodeStatus)
		if ok && cacheStatus != nil {
			oldStatus := cacheStatus.(model.NodeStatus)
			oldStatus.Status = true
			oldStatus.CPU = AGNodeStatus.CPU
			oldStatus.Mem = AGNodeStatus.Mem
			oldStatus.Disk = AGNodeStatus.Disk
			//oldStatus.Uptime=AGNodeStatus.Uptime
			global.LocalCache.Set(strconv.FormatInt(AGNodeStatus.ID, 10)+global.NodeStatus, oldStatus, 2*time.Minute) //2分钟后过期
		} else {
			var status model.NodeStatus
			status.Status = true
			status.ID = AGNodeStatus.ID
			status.CPU = AGNodeStatus.CPU
			status.Mem = AGNodeStatus.Mem
			status.Disk = AGNodeStatus.Disk
			global.LocalCache.Set(strconv.FormatInt(AGNodeStatus.ID, 10)+global.NodeStatus, status, 2*time.Minute) //2分钟后过期
		}
	})

	ctx.String(200, "success")
}

func AGGetUserlist(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	id := ctx.Query("id")
	//节点是否启用
	node, _, _ := service.CommonSqlFind[model.Node, string, model.Node](fmt.Sprintf("id = %s", id))
	if !node.Enabled {
		return
	}
	//节点属于哪些goods
	nodeIDInt, _ := strconv.ParseInt(id, 10, 64)
	goods, err := service.FindGoodsByNodeID(nodeIDInt)
	if err != nil {
		return
	}
	//goods属于哪些用户
	var goodsArr []int64
	for _, v := range goods {
		goodsArr = append(goodsArr, v.ID)
	}
	var users []model.AGUserInfo
	err = global.DB.Model(&model.User{}).Where("goods_id in (?) and sub_status = ?", goodsArr, true).Find(&users).Error
	if err != nil {
		global.Logrus.Error("error,id="+id, err.Error())
		return
	}
	//处理ss加密
	switch node.NodeType {
	case "shadowsocks":
		switch strings.HasPrefix(node.Scy, "2022") {
		case true:
			for k, _ := range users {
				p := users[k].UUID.String()
				if node.Scy == "2022-blake3-aes-128-gcm" {
					p = p[:16]
				}
				p = base64.StdEncoding.EncodeToString([]byte(p))
				users[k].Passwd = p
			}
		default:
			for k, _ := range users {
				users[k].Passwd = users[k].UUID.String()
			}
		}
	default:
	}
	EtagHandler(users, ctx)
}

func ssEncryptionHandler(node model.Node, user *model.AGUserInfo) {
	switch node.NodeType {
	case "shadowsocks":
		if strings.HasPrefix(node.Scy, "2022") {
			//
			p := user.UUID.String()
			if node.Scy == "2022-blake3-aes-128-gcm" {
				p = p[:16]
			}
			p = base64.StdEncoding.EncodeToString([]byte(p))
			user.Passwd = p

		} else {
			user.Passwd = user.UUID.String()
		}
	default:

	}
}

func AGReportUserTraffic(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	var AGUserTraffic model.AGUserTraffic
	err := ctx.ShouldBind(&AGUserTraffic)
	if err != nil {
		global.Logrus.Error("error", err.Error())
		ctx.AbortWithStatus(400)
		return
	}
	//fmt.Println("上报用户流量：", AGUserTraffic)
	//查询节点倍率
	node, _, err := service.CommonSqlFind[model.Node, string, model.Node](fmt.Sprintf("id = %d", AGUserTraffic.ID))
	if err != nil {
		global.Logrus.Error("error", err.Error())
		ctx.AbortWithStatus(400)
		return
	}
	if node.TrafficRate < 0 {
		node.TrafficRate = 1
	}
	// 处理流量统计
	var userIds []int64
	var userArr []model.User
	var trafficLog = model.TrafficLog{
		NodeID: node.ID,
	}
	//var userTrafficLog []model.UserTrafficLog
	userTrafficLogMap := make(map[int64]model.UserTrafficLog)
	for _, v := range AGUserTraffic.UserTraffic {
		//每个用户流量
		userIds = append(userIds, v.UID)
		//需要更新的用户订阅信息
		userArr = append(userArr, model.User{
			ID: v.UID,
			SubscribeInfo: model.SubscribeInfo{
				U: int64(float64(v.Upload) * node.TrafficRate),
				D: int64(float64(v.Download) * node.TrafficRate),
			},
		})
		//需要插入的用户流量统计
		userTrafficLogMap[v.UID] = model.UserTrafficLog{
			UserID:   v.UID,
			UserName: v.Email,
			U:        v.Upload,
			D:        v.Download,
		}
		//该节点总流量
		trafficLog.D = trafficLog.U + v.Upload
		trafficLog.U = trafficLog.D + v.Download

	}
	// 处理探针
	global.GoroutinePool.Submit(func() {
		var duration float64 = 60 //默认60秒间隔
		cacheStatus, ok := global.LocalCache.Get(strconv.FormatInt(AGUserTraffic.ID, 10) + global.NodeStatus)
		if ok && cacheStatus != nil {
			oldStatus := cacheStatus.(model.NodeStatus)
			oldStatus.Status = true
			oldStatus.UserAmount = int64(len(userIds))
			now := time.Now()
			duration = now.Sub(oldStatus.LastTime).Seconds()
			oldStatus.D, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(trafficLog.D)/duration), 64) //Byte per second
			oldStatus.U, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(trafficLog.U)/duration), 64)
			oldStatus.LastTime = now
			global.LocalCache.Set(strconv.FormatInt(AGUserTraffic.ID, 10)+global.NodeStatus, oldStatus, 2*time.Minute)
		} else {
			var nodeStatus model.NodeStatus
			nodeStatus.Status = true
			nodeStatus.ID = AGUserTraffic.ID
			nodeStatus.UserAmount = int64(len(userIds))
			nodeStatus.D, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(trafficLog.D)/duration), 64) //Byte per second
			nodeStatus.U, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(trafficLog.U)/duration), 64)
			nodeStatus.LastTime = time.Now()
			global.LocalCache.Set(strconv.FormatInt(AGUserTraffic.ID, 10)+global.NodeStatus, nodeStatus, 2*time.Minute)
		}
	})
	//插入节点流量统计
	global.GoroutinePool.Submit(func() {
		//查询当天的数据
		now := time.Now()
		zeroTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		traffic, _, _ := service.CommonSqlLast[model.TrafficLog, string, model.TrafficLog](fmt.Sprintf("node_id = %d AND created_at > '%v'", AGUserTraffic.ID, zeroTime))
		if traffic.ID == 0 {
			err = service.CommonSqlCreate[model.TrafficLog](trafficLog)
			if err != nil {
				global.Logrus.Error("插入流量统计统计error:", err)
				return
			}
		} else {
			traffic.U = traffic.U + trafficLog.U
			traffic.D = traffic.D + trafficLog.D
			err = service.CommonSqlSave[model.TrafficLog](traffic)
			if err != nil {
				global.Logrus.Error("插入流量统计统计error:", err)
				return
			}
		}
	})
	//插入用户流量统计
	global.GoroutinePool.Submit(func() {
		err = service.UpdateUserTrafficLog(userTrafficLogMap, userIds)
		if err != nil {
			global.Logrus.Error("插入用户流量统计,error:", err)
			return
		}
	})
	//更新用户流量信息
	global.GoroutinePool.Submit(func() {
		if len(userArr) == 0 {
			return
		}
		err = service.UpdateUserTrafficInfo(userArr, userIds)
		if err != nil {
			global.Logrus.Error("更新用户流量信息error:", err)
			return
		}
	})
	ctx.String(200, "success")

}

func AGReportNodeOnlineUsers(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	var AGOnlineUser model.AGOnlineUser
	err := ctx.ShouldBind(&AGOnlineUser)
	if err != nil {
		global.Logrus.Error("error", err.Error())
		return
	}
	ctx.String(200, "success")
}
