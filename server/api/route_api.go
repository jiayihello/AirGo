package api

import (
	"AirGo/global"
	"AirGo/model"
	"AirGo/service"
	"AirGo/utils/other_plugin"
	"AirGo/utils/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取角色动态路由
func GetRouteList(ctx *gin.Context) {
	uIdInt, _ := other_plugin.GetUserIDFromGinContext(ctx)
	//查询uId对应的角色
	roleIds, err := service.FindRoleIdsByuId(uIdInt)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("GetRouteList error:"+err.Error(), nil, ctx)
		return
	}
	// 角色Ids对应的route Ids
	routeIds, err := service.GetRouteIdsByRoleIds(roleIds)
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteIdsByRoleIds error:"+err.Error(), nil, ctx)
		return
	}
	// 根据route Ids 查 route Slice
	routeSlice, err := service.GetRouteSliceByRouteIds(routeIds)
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteSliceByRouteIds error:"+err.Error(), nil, ctx)
		return
	}
	// 获取角色动态路由
	route := service.GetDynamicRoute(routeSlice)
	response.OK("GetRouteList success", route, ctx)
}

// 获取全部角色动态路由
func GetAllRouteList(ctx *gin.Context) {
	// 根据route Ids 查 route Slice
	routeSlice, err := service.GetRouteSliceByRouteIds(nil)
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteSliceByRouteIds error:"+err.Error(), nil, ctx)
		return
	}
	// 获取角色动态路由
	route := service.GetDynamicRoute(routeSlice)
	response.OK("GetAllRouteList success", route, ctx)

}

// 前端编辑角色的时候显示全部菜单节点树
func GetAllRouteTree(ctx *gin.Context) {
	routeNodeSlice, err := service.GetRouteNodeByRouteIds(nil)
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteNodeByRouteIds error:"+err.Error(), nil, ctx)
		return
	}
	routeNodeTree := service.GetRouteNodeTree(routeNodeSlice)
	response.OK("GetAllRouteTree success", routeNodeTree, ctx)
}

// 前端编辑角色的时候显示当前角色的菜单tree
func GetRouteTree(ctx *gin.Context) {
	roleId, _ := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	// 角色Ids对应的route Ids
	var roleIds = []int64{roleId}
	routeIds, err := service.GetRouteIdsByRoleIds(roleIds) //空
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteIdsByRoleIds error:"+err.Error(), nil, ctx)
		return
	}
	routeNodeSlice, err := service.GetRouteNodeByRouteIds(routeIds)
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetRouteNodeByRouteIds error:"+err.Error(), nil, ctx)
		return
	}
	routeNodeTree := service.GetRouteNodeTree(routeNodeSlice)
	response.OK("GetRouteTree success", routeNodeTree, ctx)
}

// 新建动态路由
func NewDynamicRoute(ctx *gin.Context) {
	var route model.DynamicRoute
	err := ctx.ShouldBind(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NewDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	route.ID = 0
	// 查询动态路由是否存在
	notExist := service.NotExistDynamicRoute(&route)
	if !notExist {
		response.Fail("DynamicRoute existed", nil, ctx)
		return
	}
	err = service.NewDynamicRoute(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("NewDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("NewDynamicRoute success", nil, ctx)

}

// 删除动态路由
func DelDynamicRoute(ctx *gin.Context) {
	var route model.DynamicRoute
	err := ctx.ShouldBind(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DelDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	// 查询动态路由是否存在
	notExist := service.NotExistDynamicRoute(&route)
	if notExist {
		response.Fail("DynamicRoute does not exist", nil, ctx)
		return
	}
	err = service.DelDynamicRoute(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("DelDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("DelDynamicRoute success", nil, ctx)

}

// 修改动态路由
func UpdateDynamicRoute(ctx *gin.Context) {
	var route model.DynamicRoute
	err := ctx.ShouldBind(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("UpdateDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	// 查询动态路由是否存在
	notExist := service.NotExistDynamicRoute(&route)
	if notExist {
		response.Fail("DynamicRoute does not exist", nil, ctx)
		return
	}

	err = service.UpdateDynamicRoute(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("UpdateDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("UpdateDynamicRoute success", nil, ctx)

}

// 查询单条动态路由 by meta.title
func FindDynamicRoute(ctx *gin.Context) {
	var route model.DynamicRoute
	err := ctx.ShouldBind(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("FindDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	res, err := service.FindDynamicRoute(&route)
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail("FindDynamicRoute error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("FindDynamicRoute success", res, ctx)

}
