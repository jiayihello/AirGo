package api

import (
	"AirGo/global"
	"AirGo/model"
	"AirGo/service"
	"AirGo/utils/encrypt_plugin"
	"AirGo/utils/response"
	"github.com/gin-gonic/gin"
)

// 获取全部节点
func GetAllNode(ctx *gin.Context) {
	nodeArr, _, err := service.CommonSqlFind[model.Node, string, []model.Node]("")
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("GetAllNode error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("GetAllNode success", nodeArr, ctx)
}

// 新建节点
func NewNode(ctx *gin.Context) {
	var node model.Node
	err := ctx.ShouldBind(&node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NewNode error:"+err.Error(), nil, ctx)
		return
	}
	node.ServerKey = encrypt_plugin.RandomString(32)
	n, _, _ := service.CommonSqlFind[model.Node, model.Node, model.Node](model.Node{
		Remarks: node.Remarks,
	})
	if n.Remarks != "" {
		response.Fail("Node name is duplicate", nil, ctx)
		return
	}
	err = service.CommonSqlCreate[model.Node](node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NewNode error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("NewNode success", nil, ctx)
}

// 删除节点
func DeleteNode(ctx *gin.Context) {
	var node model.Node
	err := ctx.ShouldBind(&node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DeleteNode error:"+err.Error(), nil, ctx)
		return
	}
	err = service.CommonSqlDelete[model.Node, model.Node](node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DeleteNode error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("DeleteNode success", nil, ctx)
}

// 更新节点
func UpdateNode(ctx *gin.Context) {
	var node model.Node
	err := ctx.ShouldBind(&node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("UpdateNode error:"+err.Error(), nil, ctx)
		return
	}
	err = service.UpdateNode(&node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("UpdateNode error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("UpdateNode success", nil, ctx)

}

// 查询节点流量
func GetNodeTraffic(ctx *gin.Context) {
	var trafficParams model.PaginationParams
	err := ctx.ShouldBind(&trafficParams)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("GetNodeTraffic error:"+err.Error(), nil, ctx)
		return
	}
	res := service.GetNodeTraffic(trafficParams)
	response.OK("GetNodeTraffic success", res, ctx)
}

// 节点排序
func NodeSort(ctx *gin.Context) {
	var nodeArr []model.Node
	err := ctx.ShouldBind(&nodeArr)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NodeSort error:"+err.Error(), nil, ctx)
		return
	}
	err = service.CommonSqlUpdateMultiLine[[]model.Node](nodeArr, "id", []string{"node_order"})
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NodeSort error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("NodeSort success", nil, ctx)
}

// 新增共享节点
func NewNodeShared(ctx *gin.Context) {
	var url model.NodeSharedReq
	err := ctx.ShouldBind(&url)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NewNodeShared error:"+err.Error(), nil, ctx)
		return
	}
	nodeArr := service.ParseSubUrl(url.Url)
	if nodeArr != nil {
		for _, v := range *nodeArr {
			n, _, _ := service.CommonSqlFind[model.NodeShared, model.NodeShared, model.NodeShared](model.NodeShared{
				Remarks: v.Remarks,
			})
			if n.Remarks != "" {
				continue
			}
			err = service.CommonSqlCreate[[]model.NodeShared](*nodeArr)
			if err != nil {
				global.Logrus.Error(err.Error())
				response.Fail("NewNodeShared error:"+err.Error(), nil, ctx)
				return
			}
		}
		response.OK("NewNodeShared success", nil, ctx)
	}
}

// 获取共享节点列表
func GetNodeSharedList(ctx *gin.Context) {
	nodeArr, _, err := service.CommonSqlFind[model.NodeShared, string, []model.Node]("")
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("GetNodeSharedList"+err.Error(), nil, ctx)
		return
	}
	response.OK("GetNodeSharedList success", nodeArr, ctx)

}

// 删除共享节点
func DeleteNodeShared(ctx *gin.Context) {
	var node model.NodeShared
	err := ctx.ShouldBind(&node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DeleteNodeShared error:"+err.Error(), nil, ctx)
		return
	}
	err = service.CommonSqlDelete[model.Node, model.NodeShared](node)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DeleteNodeShared error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("DeleteNodeShared success", nil, ctx)
}
