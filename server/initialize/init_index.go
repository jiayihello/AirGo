package initialize

import (
	"context"
	"fmt"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/mojocn/base64Captcha"
	"github.com/panjf2000/ants/v2"
	"github.com/ppoonk/AirGo/global"
	"github.com/ppoonk/AirGo/model"
	"github.com/ppoonk/AirGo/service"
	"github.com/ppoonk/AirGo/utils/logrus_plugin"
	"github.com/ppoonk/AirGo/utils/mail_plugin"
	queue "github.com/ppoonk/AirGo/utils/queue_plugin"
	"github.com/ppoonk/AirGo/utils/time_plugin"
	"github.com/ppoonk/AirGo/utils/websocket_plugin"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"github.com/yudeguang/ratelimit"
	"time"
)

// 初始化全部资源，注意顺序
func InitializeAll() {
	InitLogrus()            //logrus
	global.VP = InitViper() //初始化Viper
	global.DB = Gorm()      //gorm连接数据库
	if global.DB != nil {
		if !global.DB.Migrator().HasTable(&model.User{}) {
			global.Logrus.Info("未找到sys_user库表,开始建表并初始化数据...")
			RegisterTables() //创建table
			InsertInto()     //导入数据
		} else {
			RegisterTables() //AutoMigrate 自动迁移 schema
		}
	} else {
		panic("数据库连接失败")
	}
	InitServer()        //加载全局系统配置
	InitCasbin()        //加载casbin
	InitTheme()         //加载全局主题
	InitLocalCache()    //local cache
	InitGoroutinePool() //初始化线程池

	InitBase64Captcha() //Base64Captcha
	InitEmailDialer()   //gomail Dialer
	InitWebsocket()     //websocket
	InitRatelimit()     //限流

	InitContextGroup() //
	InitTGBot()        //初始化tg bot
	InitCrontab()      //定时任务
	InitRouter()       //初始总路由，放在最后

}

// 重置管理员密码
func InitializeResetAdmin() {
	global.VP = InitViper()
	global.DB = Gorm()
	service.ResetAdminPassword()
}

// 升级核心
func InitializeUpdate() {
	global.VP = InitViper() //初始化Viper
	global.DB = Gorm()      //gorm连接数据库
	InitServer()            //加载全局系统配置

	var funcs = []func() error{
		func() error {
			fmt.Println("升级数据库casbin_rule表")
			err := global.DB.Where("id > 0").Delete(&gormadapter.CasbinRule{}).Error
			if err != nil {
				return err
			}
			return InsertIntoCasbinRule()
		},
		func() error {
			fmt.Println("升级角色和菜单")
			//先删除role_and_menu
			err := global.DB.Where("role_id > 0").Delete(&model.RoleAndMenu{}).Error
			if err != nil {
				return err
			}
			//再删除菜单
			err = global.DB.Where("id > 0").Delete(&model.DynamicRoute{}).Error
			if err != nil {
				return err
			}
			//插入新的菜单
			err = InsertIntoDynamicRoute()
			if err != nil {
				return err
			}
			//插入新的role_and_menu
			return InsertIntoRoleAndMenu()
		},
		//临时代码，处理之前版本删除节点遗留的数据库垃圾数据
		func() error {
			fmt.Println("处理遗留垃圾数据-无效节点信息")
			return service.DeleteNodeTemp()
		},
		//临时代码，删除用户流量统计
		func() error {
			fmt.Println("处理遗留垃圾数据-用户流量统计")
			return service.DeleteUserTrafficTemp()
		},
	}
	for _, v := range funcs {
		err := v()
		if err != nil {
			fmt.Println("升级核心出错：", err.Error())
			return
		}
	}
	fmt.Println("升级核心完成")

}

func InitLogrus() {
	global.Logrus = logrus_plugin.InitLogrus()
}
func InitServer() {
	//res, err := service.GetSetting()
	res, _, err := service.CommonSqlFind[model.Server, string, model.Server]("id = 1")
	if err != nil {
		global.Logrus.Error("系统配置获取失败", err.Error())
		return
	}
	global.Server = res
}
func InitCasbin() {
	global.Casbin = service.Casbin()
}
func InitTheme() {
	//res, err := service.GetThemeConfig()
	res, _, err := service.CommonSqlFind[model.Theme, string, model.Theme]("id = 1")
	if err != nil {
		global.Logrus.Error("系统配置获取失败", err.Error())
		return
	}
	global.Theme = res
}
func InitLocalCache() {
	//判断有没有设置时间
	dr := time.Hour
	if global.Server.Security.JWT.ExpiresTime != "" {
		dr, _ = time_plugin.ParseDuration(global.Server.Security.JWT.ExpiresTime)
	}
	//初始化local cache配置
	global.LocalCache = local_cache.NewCache(
		local_cache.SetDefaultExpire(dr), //设置默认的超时时间
	)
}
func InitBase64Captcha() {
	// base64Captcha.DefaultMemStore 是默认的过期时间10分钟。也可以自己设定参数 base64Captcha.NewMemoryStore(GCLimitNumber, Expiration)
	global.Base64CaptchaStore = base64Captcha.DefaultMemStore
	driver := base64Captcha.NewDriverDigit(38, 120, 4, 0.2, 10)
	global.Base64Captcha = base64Captcha.NewCaptcha(driver, global.Base64CaptchaStore)
}
func InitEmailDialer() {
	d := mail_plugin.InitEmailDialer(global.Server.Email.EmailHost, int(global.Server.Email.EmailPort), global.Server.Email.EmailFrom, global.Server.Email.EmailSecret)
	if d != nil {
		global.EmailDialer = d
	}
}
func InitWebsocket() {
	global.WsManager = websocket_plugin.NewManager()
	global.WsManager.NewClientManager()
}
func InitRatelimit() {
	global.RateLimit.IPRole = ratelimit.NewRule()
	global.RateLimit.IPRole.AddRule(time.Second*60, int(global.Server.Security.RateLimitParams.IPRoleParam))
	global.RateLimit.VisitRole = ratelimit.NewRule()
	global.RateLimit.VisitRole.AddRule(time.Second*60, int(global.Server.Security.RateLimitParams.VisitParam))
}
func InitGoroutinePool() {
	global.GoroutinePool, _ = ants.NewPool(100, ants.WithPreAlloc(true))
}
func InitContextGroup() {
	global.ContextGroup = &model.ContextGroup{
		CtxMap:    make(map[string]*context.Context),
		CancelMap: make(map[string]*context.CancelFunc),
	}
}
func InitTGBot() {
	service.TGBotStartListen()
}
func InitQueue() {
	global.Queue = queue.NewQueue()
}
