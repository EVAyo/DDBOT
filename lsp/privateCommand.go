package lsp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/youtube"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"io/ioutil"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type LspPrivateCommand struct {
	msg *message.PrivateMessage

	*Runtime
}

func NewLspPrivateCommand(bot *miraiBot.Bot, l *Lsp, msg *message.PrivateMessage) *LspPrivateCommand {
	c := &LspPrivateCommand{
		msg:     msg,
		Runtime: NewRuntime(bot, l),
	}
	c.Parse(c.msg.Elements)
	return c
}

func (c *LspPrivateCommand) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("panic recovered: %v", err)
			c.textSend("エラー発生：看到该信息表示BOT出了一些问题，该问题已记录")
		}
	}()
	if !strings.HasPrefix(c.GetCmd(), "/") {
		return
	}

	log := c.DefaultLogger().WithField("cmd", c.GetCmdArgs())

	if c.l.PermissionStateManager.CheckBlockList(c.uin()) {
		log.Debug("blocked")
		return
	}

	if !c.DebugCheck() {
		log.Debugf("debug mode, skip execute.")
		return
	}

	log.Debug("execute command")

	// all permission will be checked later
	switch c.GetCmd() {
	case "/ping":
		c.PingCommand()
	case "/help":
		c.HelpCommand()
	case "/block":
		c.BlockCommand()
	case "/watch":
		c.WatchCommand(false)
	case "/unwatch":
		c.WatchCommand(true)
	case "/enable":
		c.EnableCommand(false)
	case "/disable":
		c.EnableCommand(true)
	case "/grant":
		c.GrantCommand()
	case "/log":
		c.LogCommand()
	case "/list":
		c.ListCommand()
	case "/sysinfo":
		c.SysinfoCommand()
	case "/config":
		c.ConfigCommand()
	case "/whosyourdaddy":
		c.WhosyourdaddyCommand()
	case "/quit":
		c.QuitCommand()
	case "/mode":
		c.ModeCommand()
	case "/群邀请":
		c.GroupRequestCommand()
	case "/好友申请":
		c.FriendRequestCommand()
	case "/admin":
		c.AdminCommand()
	case "/silence":
		c.SilenceCommand()
	default:
		c.textReply("阁下似乎输入了一个无法识别的命令，请使用/help命令查看帮助。")
		log.Debug("no command matched")
	}
}

func (c *LspPrivateCommand) WhosyourdaddyCommand() {
	log := c.DefaultLoggerWithCommand(WhosyourdaddyCommand)
	log.Info("run whosyourdaddy command")
	defer func() { log.Info("whosyourdaddy command end") }()

	_, output := c.parseCommandSyntax(&struct{}{}, WhosyourdaddyCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if c.l.PermissionStateManager.CheckRole(c.uin(), permission.Admin) {
		c.textReply("您已经是管理员了，请不要重复使用此命令。")
		return
	}

	if c.l.PermissionStateManager.CheckNoAdmin() {
		if err := c.l.PermissionStateManager.GrantRole(c.uin(), permission.Admin); err != nil {
			log.WithField("permission", permission.Admin.String()).
				Errorf("GrantRole error %v", err)
			c.textReply("失败 - 内部错误")
		} else {
			log.Info("you are the admin now, good luck!")
			c.textReply("成功 - 您已成为bot管理员")
		}
	} else {
		log.Debug("someone is trying WhosyourdaddyCommand")
		c.textReply("失败 - 该bot不属于你！")
	}
}

func (c *LspPrivateCommand) ListCommand() {
	log := c.DefaultLoggerWithCommand(ListCommand)
	log.Info("run list command")
	defer func() { log.Info("list command end") }()

	var listCmd struct {
		Site  string `optional:"" short:"s" help:"已弃用"`
		Type  string `optional:"" short:"t" help:"已弃用"`
		Group int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
	}
	_, output := c.parseCommandSyntax(&listCmd, ListCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if listCmd.Site != "" || listCmd.Type != "" {
		c.textReply("命令已更新，请直接输入/list即可")
		return
	}

	groupCode := listCmd.Group
	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}
	log = log.WithFields(localutils.GroupLogFields(groupCode))
	IList(c.NewMessageContext(log), groupCode)
}

func (c *LspPrivateCommand) ConfigCommand() {
	log := c.DefaultLoggerWithCommand(ConfigCommand)
	log.Info("run config command")
	defer func() { log.Info("config command end") }()

	var configCmd struct {
		At struct {
			Site   string  `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string  `arg:"" help:"配置的主播id"`
			Action string  `arg:"" enum:"add,remove,clear,show" help:"add / remove / clear / show"`
			QQ     []int64 `arg:"" optional:"" help:"需要@的成员QQ号码"`
		} `cmd:"" help:"配置推送时的@人员列表" name:"at"`
		AtAll struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"on" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置推送时@全体成员，需要管理员权限" name:"at_all"`
		TitleNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置直播间标题发生变化时是否进行推送，默认不推送" name:"title_notify"`
		OfflineNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off," help:"on / off"`
		} `cmd:"" help:"配置下播时是否进行推送，默认不推送" name:"offline_notify"`
		Filter struct {
			Site string `optional:"" short:"s" default:"bilibili" help:"bilibili"`
			Type struct {
				Id   string   `arg:"" help:"配置的主播id"`
				Type []string `arg:"" optional:"" help:"指定的种类"`
			} `cmd:"" help:"只推送指定种类的动态" name:"type" group:"filter"`
			NotType struct {
				Id   string   `arg:"" help:"配置的主播id"`
				Type []string `arg:"" optional:"" help:"指定不推送的种类"`
			} `cmd:"" help:"不推送指定种类的动态" name:"not_type" group:"filter"`
			Text struct {
				Id      string   `arg:"" help:"配置的主播id"`
				Keyword []string `arg:"" optional:"" help:"指定的关键字"`
			} `cmd:"" help:"当动态内容里出现关键字时进行推送" name:"text" group:"filter"`
			Clear struct {
				Id string `arg:"" help:"配置的主播id"`
			} `cmd:"" help:"清除过滤器" name:"clear" group:"filter"`
			Show struct {
				Id string `arg:"" help:"配置的主播id"`
			} `cmd:"" help:"查看当前过滤器" name:"show" group:"filter"`
		} `cmd:"" help:"配置动态过滤器，目前只支持b站动态" name:"filter"`
		Group int64 `optional:"" short:"g" help:"要操作的QQ群号码"`
	}

	kongCtx, output := c.parseCommandSyntax(&configCmd, ConfigCommand, kong.Description("管理BOT的配置，目前支持配置@成员、@全体成员、开启下播推送、开启标题推送"))
	if output != "" {
		c.textReply(output)
	}
	if c.exit || len(kongCtx.Path) <= 1 {
		return
	}

	groupCode := configCmd.Group
	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}

	kongPath := strings.Split(kongCtx.Command(), " ")

	cmd := kongPath[0]
	log = log.WithFields(localutils.GroupLogFields(groupCode)).WithField("sub_command", cmd)

	switch cmd {
	case "at":
		site, ctype, err := c.ParseRawSiteAndType(configCmd.At.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.At.Site).Errorf("ParseRawSiteAndType failed %v", err)
			c.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		log = log.WithField("site", site).WithField("id", configCmd.At.Id).WithField("action", configCmd.At.Action).WithField("QQ", configCmd.At.QQ)
		IConfigAtCmd(c.NewMessageContext(log), groupCode, configCmd.At.Id, site, ctype, configCmd.At.Action, configCmd.At.QQ)
	case "at_all":
		site, ctype, err := c.ParseRawSiteAndType(configCmd.AtAll.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.AtAll.Site).Errorf("ParseRawSiteAndType failed %v", err)
			c.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		var on = localutils.Switch2Bool(configCmd.AtAll.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.AtAll.Id).WithField("on", on)
		IConfigAtAllCmd(c.NewMessageContext(log), groupCode, configCmd.AtAll.Id, site, ctype, on)
	case "title_notify":
		site, ctype, err := c.ParseRawSiteAndType(configCmd.TitleNotify.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.TitleNotify.Site).Errorf("ParseRawSiteAndType failed %v", err)
			c.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		var on = localutils.Switch2Bool(configCmd.TitleNotify.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.TitleNotify.Id).WithField("on", on)
		IConfigTitleNotifyCmd(c.NewMessageContext(log), groupCode, configCmd.TitleNotify.Id, site, ctype, on)
	case "offline_notify":
		site, ctype, err := c.ParseRawSiteAndType(configCmd.OfflineNotify.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.OfflineNotify.Site).Errorf("ParseRawSiteAndType failed %v", err)
			c.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		if site == youtube.Site {
			log.WithField("site", configCmd.OfflineNotify.Site).Errorf("not supported")
			c.textSend(fmt.Sprintf("失败 - %v", "暂不支持YTB"))
			return
		}
		var on = localutils.Switch2Bool(configCmd.OfflineNotify.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.OfflineNotify.Id).WithField("on", on)
		IConfigOfflineNotifyCmd(c.NewMessageContext(log), groupCode, configCmd.OfflineNotify.Id, site, ctype, on)
	case "filter":
		filterCmd := kongPath[1]
		site, ctype, err := c.ParseRawSiteAndType(configCmd.Filter.Site, "news")
		if err != nil {
			log.WithField("site", configCmd.Filter.Site).Errorf("ParseRawSiteAndType failed %v", err)
			c.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		switch filterCmd {
		case "type":
			IConfigFilterCmdType(c.NewMessageContext(log), groupCode, configCmd.Filter.Type.Id, site, ctype, configCmd.Filter.Type.Type)
		case "not_type":
			IConfigFilterCmdNotType(c.NewMessageContext(log), groupCode, configCmd.Filter.NotType.Id, site, ctype, configCmd.Filter.NotType.Type)
		case "text":
			IConfigFilterCmdText(c.NewMessageContext(log), groupCode, configCmd.Filter.Text.Id, site, ctype, configCmd.Filter.Text.Keyword)
		case "clear":
			IConfigFilterCmdClear(c.NewMessageContext(log), groupCode, configCmd.Filter.Clear.Id, site, ctype)
		case "show":
			IConfigFilterCmdShow(c.NewMessageContext(log), groupCode, configCmd.Filter.Show.Id, site, ctype)
		default:
			log.WithField("filter_cmd", filterCmd).Errorf("unknown filter command")
			c.textSend("未知的filter子命令")
		}
	default:
		c.textSend("暂未支持，你可以催作者GKD")
	}

}

func (c *LspPrivateCommand) WatchCommand(remove bool) {
	log := c.DefaultLoggerWithCommand(WatchCommand).WithField("unwatch", remove)
	log.Info("run watch command")
	defer func() { log.Info("watch command end") }()

	var (
		site      = bilibili.Site
		watchType = concern.BibiliLive
		err       error
	)

	var name string
	if remove {
		name = "unwatch"
	} else {
		name = "watch"
	}

	var watchCmd struct {
		Site  string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
		Type  string `optional:"" short:"t" default:"live" help:"news / live"`
		Group int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
		Id    string `arg:""`
	}

	_, output := c.parseCommandSyntax(&watchCmd, name)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	site, watchType, err = c.ParseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log = log.WithField("args", c.GetArgs())
		log.Errorf("parse raw concern failed %v", err)
		c.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id
	groupCode := watchCmd.Group

	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}

	log = log.WithFields(localutils.GroupLogFields(groupCode))

	IWatch(c.NewMessageContext(log), groupCode, id, site, watchType, remove)
}

func (c *LspPrivateCommand) EnableCommand(disable bool) {
	log := c.DefaultLoggerWithCommand(EnableCommand).WithField("disable", disable)
	log.Info("run enable command")
	defer func() { log.Info("enable command end") }()

	var name string
	if disable {
		name = "disable"
	} else {
		name = "enable"
	}

	var enableCmd struct {
		Group   int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
		Command string `arg:"" optional:"" help:"命令名"`
		Global  bool   `optional:"" help:"系统级操作，对所有群生效"`
	}

	_, output := c.parseCommandSyntax(&enableCmd, name)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if len(enableCmd.Command) == 0 {
		c.textReply("失败 - 没有指定要操作的命令名")
		log.Errorf("empty command")
		return
	}

	command := CombineCommand(enableCmd.Command)
	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.textReply("失败 - 命令名非法")
		return
	}

	if enableCmd.Global {
		if !c.l.PermissionStateManager.CheckRole(c.uin(), permission.Admin) {
			c.noPermission()
			return
		}
		if enableCmd.Group != 0 {
			c.textReply(fmt.Sprintf("注意：--global模式，忽略参数-g %v", enableCmd.Group))
		}

		var err error

		if disable {
			err = c.l.PermissionStateManager.GlobalDisableGroupCommand(command)
		} else {
			err = c.l.PermissionStateManager.GlobalEnableGroupCommand(command)
		}
		if err == nil {
			c.textReply("成功")
		} else if err == permission.ErrPermissionExist {
			if disable {
				c.textReply("失败 - 该命令已禁用")
			} else {
				c.textReply("失败 - 该命令已启用")
			}
		}
	} else {
		if c.l.PermissionStateManager.CheckGlobalCommandDisabled(command) {
			c.globalDisabledReply()
			return
		}

		groupCode := enableCmd.Group
		if err := c.checkGroupCode(groupCode); err != nil {
			c.textReply(err.Error())
			return
		}

		log = log.WithFields(localutils.GroupLogFields(groupCode)).
			WithField("global", enableCmd.Global).
			WithField("targetCommand", enableCmd.Command)

		IEnable(c.NewMessageContext(log), groupCode, command, disable)
	}
}

func (c *LspPrivateCommand) GrantCommand() {
	log := c.DefaultLoggerWithCommand(GrantCommand)
	log.Info("run grant command")
	defer func() { log.Info("grant command end") }()

	var grantCmd struct {
		Group   int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
		Command string `optional:"" short:"c" xor:"1" help:"command name"`
		Role    string `optional:"" short:"r" xor:"1" enum:"Admin,GroupAdmin," help:"Admin / GroupAdmin"`
		Delete  bool   `short:"d" help:"perform a ungrant instead"`
		Target  int64  `arg:""`
	}
	_, output := c.parseCommandSyntax(&grantCmd, GrantCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	groupCode := grantCmd.Group

	grantFrom := c.uin()
	grantTo := grantCmd.Target
	if grantCmd.Command == "" && grantCmd.Role == "" {
		log.Errorf("command and role both empty")
		c.textReply("参数错误 - 必须指定-c / -r")
		return
	}

	del := grantCmd.Delete
	log = log.WithField("grantFrom", grantFrom).WithField("grantTo", grantTo).WithField("delete", del)

	if grantCmd.Command != "" {
		if err := c.checkGroupCode(groupCode); err != nil {
			c.textReply(err.Error())
			return
		}
		log = log.WithFields(localutils.GroupLogFields(groupCode))
		IGrantCmd(c.NewMessageContext(log), groupCode, grantCmd.Command, grantTo, del)
	} else if grantCmd.Role != "" {
		role := permission.NewRoleFromString(grantCmd.Role)
		if role != permission.Admin {
			if err := c.checkGroupCode(groupCode); err != nil {
				c.textReply(err.Error())
				return
			}
		}
		log = log.WithFields(localutils.GroupLogFields(groupCode))
		IGrantRole(c.NewMessageContext(log), groupCode, role, grantTo, del)
	}
}

func (c *LspPrivateCommand) BlockCommand() {
	log := c.DefaultLoggerWithCommand(BlockCommand)
	log.Info("run block command")
	defer func() { log.Info("block command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var blockCmd struct {
		Uin    int64 `arg:"" required:"" help:"the uin to block"`
		Days   int   `optional:""`
		Delete bool  `optional:"" short:"d"`
	}

	_, output := c.parseCommandSyntax(&blockCmd, BlockCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if blockCmd.Uin == c.uin() {
		log.Errorf("can not block yourself")
		c.textReply("失败 - 不能block自己")
		return
	}

	if blockCmd.Days == 0 {
		blockCmd.Days = 7
	}

	log = log.WithField("target", blockCmd.Uin).WithField("days", blockCmd.Days).WithField("delete", blockCmd.Delete)

	if !blockCmd.Delete {
		if err := c.l.PermissionStateManager.AddBlockList(blockCmd.Uin, time.Duration(blockCmd.Days)*time.Hour*24); err == nil {
			log.Info("blocked")
			c.textReply("成功")
		} else if err == localdb.ErrKeyExist {
			log.Errorf("block failed - duplicate")
			c.textReply("失败 - 已经block过了")
		} else {
			log.Errorf("block failed err %v", err)
			c.textReply("失败")
		}
	} else {
		if err := c.l.PermissionStateManager.DeleteBlockList(blockCmd.Uin); err == nil {
			log.Info("unblocked")
			c.textReply("成功")
		} else if err == buntdb.ErrNotFound {
			log.Errorf("unblock failed - not exist")
			c.textReply("失败 - 该id未被block")
		} else {
			log.Errorf("unblock failed err %v", err)
			c.textReply("失败")
		}
	}
}

func (c *LspPrivateCommand) LogCommand() {
	log := c.DefaultLoggerWithCommand(LogCommand)
	log.Info("run log command")
	defer func() { log.Info("log command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var logCmd struct {
		N       int       `arg:"" optional:"" help:"the number of lines from tail"`
		Date    time.Time `optional:"" short:"d" format:"2006-01-02"`
		Keyword string    `optional:"" short:"k" help:"the lines contains at lease one keyword"`
	}

	_, output := c.parseCommandSyntax(&logCmd, LogCommand)
	if output != "" {
		c.textSend(output)
	}
	if c.exit {
		return
	}

	if logCmd.N == 0 {
		logCmd.N = 10
	}
	if logCmd.Date.IsZero() {
		logCmd.Date = time.Now()
	}
	logName := fmt.Sprintf("%v.log", logCmd.Date.Format("2006-01-02"))
	b, err := ioutil.ReadFile("logs/" + logName)
	if err != nil {
		c.textSend(fmt.Sprintf("失败 - %v", err))
		return
	}
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	if len(logCmd.Keyword) != 0 {
		var filteredLines []string
		for _, line := range lines {
			if strings.Contains(line, logCmd.Keyword) {
				filteredLines = append(filteredLines, line)
			}
		}
		lines = filteredLines
	}

	if logCmd.N > len(lines) {
		logCmd.N = len(lines)
	}

	lines = lines[len(lines)-logCmd.N:]

	if len(lines) == 0 {
		c.textSend("无结果")
	} else {
		c.textSend(strings.Join(lines, "\n"))
	}
}

func (c *LspPrivateCommand) QuitCommand() {
	log := c.DefaultLoggerWithCommand(QuitCommand)
	log.Info("run quit command")
	defer func() { log.Info("quit command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var quitCmd struct {
		GroupCode int64 `arg:"" optional:"" help:"要退出的群号"`
		Force     bool  `optional:"" short:"f" help:"强制清除"`
	}

	_, output := c.parseCommandSyntax(&quitCmd, QuitCommand)
	if output != "" {
		c.textSend(output)
	}
	if c.exit {
		return
	}

	if quitCmd.GroupCode == 0 {
		log.Errorf("GroupCode is 0")
		c.textSend("没有指定群号，请输入群号")
		return
	}

	log = log.WithField("TargetGroupCode", quitCmd.GroupCode).WithField("Force", quitCmd.Force)

	gi := c.bot.FindGroup(quitCmd.GroupCode)
	var displayName string
	if gi == nil {
		displayName = strconv.FormatInt(quitCmd.GroupCode, 10)
	} else {
		displayName = gi.Name
	}
	if gi == nil {
		if quitCmd.Force {
			log.Debugf("没有找到该QQ群，force已启用")
			c.textSend("没有找到该QQ群，请确认bot在群内，但由于指定了-f参数，将强制清除bot在该群的数据")
		} else {
			log.Errorf("没有找到该QQ群，force已禁用")
			c.textSend("没有找到该QQ群，请确认bot在群内，如果要强制清除bot在该群内的数据，请指定-f参数")
			return
		}
	} else {
		gi.Quit()
		log.Debugf("已退出群【%v】", displayName)
		c.textSend(fmt.Sprintf("已退出群【%v】", displayName))
	}
	c.l.RemoveAllByGroup(quitCmd.GroupCode)
	log.Debugf("已清除群【%v】的数据", displayName)
	c.textSend(fmt.Sprintf("已清除群【%v】的数据", displayName))
}

func (c *LspPrivateCommand) ModeCommand() {
	log := c.DefaultLoggerWithCommand(ModeCommand)
	log.Info("run mode command")
	defer func() { log.Info("mode command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var modeCmd struct {
		Mode string `arg:"" optional:"" help:"指定切换模式，支持<公开> <私人> <审核>，默认为公开"`
	}

	_, output := c.parseCommandSyntax(&modeCmd, ModeCommand, kong.Description("切换BOT模式"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if modeCmd.Mode == "" {
		mode := c.l.LspStateManager.GetCurrentMode()
		log.Infof("当前Mode为%v", mode)
		switch c.l.LspStateManager.GetCurrentMode() {
		case PublicMode:
			c.textReply("当前模式为公开")
		case PrivateMode:
			c.textReply("当前模式为私人")
		case ProtectMode:
			c.textReply("当前模式为审核")
		}
		return
	}

	log = log.WithField("Mode", modeCmd.Mode)

	var err error

	switch modeCmd.Mode {
	case "公开":
		err = c.l.LspStateManager.SetMode(PublicMode)
	case "私人":
		err = c.l.LspStateManager.SetMode(PrivateMode)
	case "审核":
		err = c.l.LspStateManager.SetMode(ProtectMode)
	default:
		log.Errorf("未知的模式")
		c.textSend(fmt.Sprintf("未知的模式【%v】，仅支持<公开> <私人> <审核>，请查看命令文档", modeCmd.Mode))
		return
	}
	if err != nil {
		log.Errorf("切换模式失败 %v", err)
		c.textReply(fmt.Sprintf("切换模式失败 - %v", err))
	} else {
		log.Infof("切换到%v模式", modeCmd.Mode)
		c.textReply(fmt.Sprintf("成功 - 切换到%v模式", modeCmd.Mode))
	}
}

func (c *LspPrivateCommand) GroupRequestCommand() {
	log := c.DefaultLoggerWithCommand(GroupRequestCommand)
	log.Info("run GroupRequest command")
	defer func() { log.Info("GroupRequest command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var groupRequestCmd struct {
		RequestId int64    `arg:"" optional:"" help:"要处理的请求的RequestId"`
		Reject    bool     `optional:"" short:"r" help:"拒绝请求"`
		All       bool     `optional:"" short:"a" help:"处理全部"`
		Message   []string `arg:"" optional:"" help:"拒绝理由"`
	}

	_, output := c.parseCommandSyntax(&groupRequestCmd, GroupRequestCommand, kong.Description("处理群邀请"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	msg := strings.Join(groupRequestCmd.Message, " ")

	log = log.WithFields(logrus.Fields{
		"Requestid": groupRequestCmd.RequestId,
		"Reject":    groupRequestCmd.Reject,
		"All":       groupRequestCmd.All,
		"Message":   msg,
	})

	if groupRequestCmd.RequestId == 0 {
		requests, err := c.l.LspStateManager.ListGroupInvitedRequest()
		if err != nil {
			log.Errorf("ListGroupInvitedRequest error - %v", err)
			c.textReply(fmt.Sprintf("失败 - %v", err))
			return
		}

		if len(requests) == 0 {
			log.Infof("没有查询到加群邀请")
			c.textReply("没有查询到加群邀请")
			return
		}

		if groupRequestCmd.Reject {
			// 拒绝全部？
			if !groupRequestCmd.All {
				log.Info("拒绝全部加群邀请需要确认")
				c.textReply(fmt.Sprintf("您似乎想要拒绝全部%v个加群邀请，如果确定想这样做，请输入：\n%v --all %v", len(requests), c.GetCmd(), strings.Join(c.GetArgs(), " ")))
				return
			}
			// 拒绝全部！
			log.Info("确认拒绝全部加群邀请")
			for _, req := range requests {
				log.Debugf("正在拒绝%v(%v)的加群%v(%v)邀请", req.InvitorNick, req.InvitorUin, req.GroupName, req.GroupCode)
				c.bot.SolveGroupJoinRequest(req, false, false, msg)
				if err := c.l.LspStateManager.DeleteGroupInvitedRequest(req.RequestId); err != nil {
					log.Errorf("DeleteGroupInvitedRequest error %v", err)
				}
			}
			log.Infof("已拒绝%v个加群邀请", len(requests))
			c.textReply(fmt.Sprintf("成功 - 已拒绝全部%v个加群邀请", len(requests)))
			return
		}

		if groupRequestCmd.All {
			// 接受全部！
			for _, req := range requests {
				log.Infof("正在接受%v(%v)的加群%v(%v)邀请", req.InvitorNick, req.InvitorUin, req.GroupName, req.GroupCode)
				c.bot.SolveGroupJoinRequest(req, true, false, "")
				if err := c.l.LspStateManager.DeleteGroupInvitedRequest(req.RequestId); err != nil {
					log.Errorf("DeleteGroupInvitedRequest error %v", err)
				}
				if err := c.l.PermissionStateManager.GrantGroupRole(req.GroupCode, req.InvitorUin, permission.GroupAdmin); err != nil {
					log.Errorf("设置群管理员权限失败 - %v", err)
				}
			}
			log.Infof("已接受%v个加群邀请", len(requests))
			c.textReply(fmt.Sprintf("成功 - 已接受全部%v个加群邀请", len(requests)))
			return
		}

		// 展示加群邀请
		var sb strings.Builder
		for _, req := range requests {
			sb.WriteString(fmt.Sprintf("ID:%v %v(%v) 邀请加入群 %v(%v)\n", req.RequestId, req.InvitorNick, req.InvitorUin, req.GroupName, req.GroupCode))
		}
		log.Infof("查询到%v个加群邀请", len(requests))
		c.textReply(sb.String())
	} else {
		request, err := c.l.LspStateManager.GetGroupInvitedRequest(groupRequestCmd.RequestId)
		if err == buntdb.ErrNotFound {
			log.Errorf("处理加群邀请失败 - 未找到该邀请")
			c.textReply(fmt.Sprintf("失败 - 未找到该邀请【%v】", groupRequestCmd.RequestId))
			return
		} else if err != nil {
			log.Errorf("GetGroupInvitedRequest error %v", err)
			c.textReply(fmt.Sprintf("失败 - 内部错误"))
			return
		}
		log := log.WithFields(logrus.Fields{
			"GroupName":   request.GroupName,
			"GroupCode":   request.GroupCode,
			"InvitorUin":  request.InvitorUin,
			"InvitorNick": request.InvitorNick,
		})
		if groupRequestCmd.Reject {
			c.bot.SolveGroupJoinRequest(request, false, false, msg)
			log.Info("拒绝加群邀请成功")
			c.textReply(fmt.Sprintf("成功- 已拒绝 %v(%v) 邀请加群 %v(%v)", request.InvitorNick, request.InvitorUin, request.GroupName, request.GroupCode))
		} else {
			c.bot.SolveGroupJoinRequest(request, true, false, "")
			if err := c.l.PermissionStateManager.GrantGroupRole(request.GroupCode, request.InvitorUin, permission.GroupAdmin); err != nil {
				log.Errorf("设置群管理员权限失败 - %v", err)
			}
			log.Info("接受加群请求成功")
			c.textReply(fmt.Sprintf("成功 - 已接受 %v(%v) 邀请加群 %v(%v)", request.InvitorNick, request.InvitorUin, request.GroupName, request.GroupCode))
		}
		if err := c.l.LspStateManager.DeleteGroupInvitedRequest(request.RequestId); err != nil {
			log.Errorf("DeleteGroupInvitedRequest error %v", err)
		}
	}
}

func (c *LspPrivateCommand) FriendRequestCommand() {
	log := c.DefaultLoggerWithCommand(FriendRequestCommand)
	log.Info("run FriendRequestCommand command")
	defer func() { log.Info("FriendRequestCommand command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var friendRequestCmd struct {
		RequestId int64 `arg:"" optional:"" help:"要处理的请求的RequestId"`
		Reject    bool  `optional:"" short:"r" help:"拒绝请求"`
		All       bool  `optional:"" short:"a" help:"处理全部"`
	}

	_, output := c.parseCommandSyntax(&friendRequestCmd, FriendRequestCommand, kong.Description("处理好友请求"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	log = log.WithFields(logrus.Fields{
		"Requestid": friendRequestCmd.RequestId,
		"Reject":    friendRequestCmd.Reject,
		"All":       friendRequestCmd.All,
	})

	if friendRequestCmd.RequestId == 0 {
		requests, err := c.l.LspStateManager.ListNewFriendRequest()
		if err != nil {
			log.Errorf("ListNewFriendRequest error - %v", err)
			c.textReply(fmt.Sprintf("失败 - %v", err))
			return
		}

		if len(requests) == 0 {
			log.Infof("没有查询到好友申请")
			c.textReply("没有查询到好友申请")
			return
		}

		if friendRequestCmd.Reject {
			if !friendRequestCmd.All {
				log.Info("拒绝全部好友申请需要确认")
				c.textReply(fmt.Sprintf("您似乎想要拒绝全部%v个好友申请，如果确定想这样做，请输入：\n%v --all %v", len(requests), c.GetCmd(), strings.Join(c.GetArgs(), " ")))
				return
			}
			log.Info("确认拒绝全部好友申请")
			for _, req := range requests {
				log.Debugf("正在拒绝%v(%v)的好友申请", req.RequesterNick, req.RequesterUin)
				c.bot.SolveFriendRequest(req, false)
				if err := c.l.LspStateManager.DeleteNewFriendRequest(req.RequestId); err != nil {
					log.Errorf("DeleteNewFriendRequest error %v", err)
				}
			}
			log.Infof("已拒绝%v个好友申请", len(requests))
			c.textReply(fmt.Sprintf("成功 - 已拒绝全部%v个好友申请", len(requests)))
			return
		}

		if friendRequestCmd.All {
			// 接受全部！
			for _, req := range requests {
				log.Debugf("正在接受%v(%v)的好友申请", req.RequesterNick, req.RequesterUin)
				c.bot.SolveFriendRequest(req, true)
				if err := c.l.LspStateManager.DeleteNewFriendRequest(req.RequestId); err != nil {
					log.Errorf("DeleteNewFriendRequest error %v", err)
				}
			}
			log.Infof("已接受%v个好友申请", len(requests))
			c.textReply(fmt.Sprintf("成功 - 已接受全部%v个好友申请", len(requests)))
			return
		}

		// 展示好友申请
		var sb strings.Builder
		for _, req := range requests {
			sb.WriteString(fmt.Sprintf("ID:%v %v(%v)申请好友\n", req.RequestId, req.RequesterNick, req.RequesterUin))
		}
		log.Infof("查询到%v个好友申请", len(requests))
		c.textReply(sb.String())
	} else {
		request, err := c.l.LspStateManager.GetNewFriendRequest(friendRequestCmd.RequestId)
		if err == buntdb.ErrNotFound {
			log.Errorf("处理好友申请失败 - 未找到该好友申请")
			c.textReply(fmt.Sprintf("失败 - 未找到该好友申请【%v】", friendRequestCmd.RequestId))
			return
		} else if err != nil {
			log.Errorf("GetNewFriendRequest error %v", err)
			c.textReply(fmt.Sprintf("失败 - 内部错误"))
			return
		}

		log := log.WithFields(logrus.Fields{
			"RequesterNick": request.RequesterNick,
			"RequesterUin":  request.RequesterUin,
		})

		if friendRequestCmd.Reject {
			c.bot.SolveFriendRequest(request, false)
			log.Info("拒绝好友申请")
			c.textReply(fmt.Sprintf("成功 - 已拒绝 %v(%v) 的好友申请", request.RequesterNick, request.RequesterUin))
		} else {
			c.bot.SolveFriendRequest(request, true)
			log.Info("接受好友申请")
			c.textReply(fmt.Sprintf("成功 - 已接受 %v(%v) 的好友申请", request.RequesterNick, request.RequesterUin))
		}
		if err := c.l.LspStateManager.DeleteNewFriendRequest(request.RequestId); err != nil {
			log.Errorf("DeleteNewFriendRequest error %v", err)
		}
	}
}

func (c *LspPrivateCommand) AdminCommand() {
	log := c.DefaultLoggerWithCommand(AdminCommand)
	log.Info("run admin command")
	defer func() { log.Info("admin command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	_, output := c.parseCommandSyntax(&struct{}{}, AdminCommand, kong.Description("查看当前Admin权限"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	ids := c.l.PermissionStateManager.ListAdmin()
	if len(ids) == 0 {
		c.textReply("未查询到Admin，如果bot刚刚启动，请稍后重试。")
		return
	}
	msg := message.NewSendingMessage()
	msg.Append(message.NewText("当前Admin："))
	var name string
	var sb strings.Builder
	for _, id := range ids {
		fi := c.bot.FindFriend(id)
		if fi != nil {
			name = fi.Nickname
		}
		sb.WriteString(fmt.Sprintf("\n%v %v", id, name))
	}
	msg.Append(message.NewText(sb.String()))
	c.send(msg)
}

func (c *LspPrivateCommand) SilenceCommand() {
	log := c.DefaultLoggerWithCommand(SilenceCommand)
	log.Info("run silence command")
	defer func() { log.Info("silence command end") }()

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
	) {
		c.noPermission()
		return
	}

	var silenceCmd struct {
		Group  int64 `optional:"" short:"g" help:"要操作的QQ群号码"`
		Delete bool  `optional:"" short:"d" help:"取消设置"`
	}

	_, output := c.parseCommandSyntax(&silenceCmd, SilenceCommand, kong.Description("设置沉默模式"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	ISilenceCmd(c.NewMessageContext(log), silenceCmd.Group, silenceCmd.Delete)
}

func (c *LspPrivateCommand) PingCommand() {
	log := c.DefaultLoggerWithCommand(PingCommand)
	log.Info("run ping command")
	defer func() { log.Info("ping command end") }()

	_, output := c.parseCommandSyntax(&struct{}{}, PingCommand, kong.Description("reply a pong"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}
	c.textReply("pong")
}

func (c *LspPrivateCommand) HelpCommand() {
	log := c.DefaultLoggerWithCommand(HelpCommand)
	log.Info("run help command")
	defer func() { log.Info("help command end") }()

	_, output := c.parseCommandSyntax(&struct{}{}, HelpCommand, kong.Description("print help message"))
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	help := "常见订阅用法：\n" +
		"以作者UID:97505为例\n" +
		"首先订阅直播信息：/watch 97505\n" +
		"然后订阅动态信息：/watch -t news 97505\n" +
		"由于通常动态内容较多，可以选择不推送转发的动态\n" +
		"/config filter not_type 97505 转发\n" +
		"还可以选择开启直播推送时@全体成员：\n" +
		"/config at_all 97505 on\n" +
		"以及开启下播推送：\n" +
		"/config offline_notify 97505 on\n" +
		"BOT还支持更多功能，详细命令介绍请查看命令文档：\n" +
		"https://hub.fastgit.org/Sora233/DDBOT/blob/master/EXAMPLE.md\n" +
		"使用时请把作者UID换成你需要的UID\n" +
		"当您完成所有配置后，可以使用/silence命令，让bot专注于推送，在群内发言更少"
	help2 := "B站专栏介绍：https://www.bilibili.com/read/cv10602230\n" +
		"如果您有任何疑问或者建议，请反馈到唯一指定交流群：755612788"
	c.textSend(help)
	time.AfterFunc(time.Millisecond*500, func() {
		c.textReply(help2)
	})
}

func (c *LspPrivateCommand) SysinfoCommand() {
	log := c.DefaultLoggerWithCommand(SysinfoCommand)
	log.Info("run sysinfo command")
	defer func() { log.Info("sysinfo command end") }()

	_, output := c.parseCommandSyntax(&struct{}{}, SysinfoCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if !c.l.PermissionStateManager.RequireAny(permission.AdminRoleRequireOption(c.uin())) {
		c.noPermission()
		return
	}

	if c.bot == nil || !c.l.started {
		c.textReply("当前暂时无法查询")
		return
	}

	msg := message.NewSendingMessage()
	msg.Append(localutils.MessageTextf("当前好友数：%v\n", len(c.bot.FriendList)))
	msg.Append(localutils.MessageTextf("当前群组数：%v\n", len(c.bot.GroupList)))
	ids, err := c.l.bilibiliConcern.ListIds()
	if err != nil {
		msg.Append(localutils.MessageTextf("当前Bilibili订阅数：获取失败\n"))
	} else {
		msg.Append(localutils.MessageTextf("当前Bilibili订阅数：%v\n", len(ids)))
	}
	ids, err = c.l.douyuConcern.ListIds()
	if err != nil {
		msg.Append(localutils.MessageTextf("当前Douyu订阅数：获取失败\n"))
	} else {
		msg.Append(localutils.MessageTextf("当前Douyu订阅数：%v\n", len(ids)))
	}
	ids, err = c.l.youtubeConcern.ListIds()
	if err != nil {
		msg.Append(localutils.MessageTextf("当前YTB订阅数：获取失败\n"))
	} else {
		msg.Append(localutils.MessageTextf("当前YTB订阅数：%v\n", len(ids)))
	}
	ids, err = c.l.huyaConcern.ListIds()
	if err != nil {
		msg.Append(localutils.MessageTextf("当前Huya订阅数：获取失败\n"))
	} else {
		msg.Append(localutils.MessageTextf("当前Huya订阅数：%v\n", len(ids)))
	}
	c.send(msg)
}

func (c *LspPrivateCommand) DebugCheck() bool {
	var ok bool
	if c.debug {
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.uin"), c.msg.Sender) {
			ok = true
		}
	} else {
		ok = true
	}
	return ok
}

func (c *LspPrivateCommand) DefaultLogger() *logrus.Entry {
	return logger.WithField("Uin", c.uin()).WithField("Name", c.name())
}

func (c *LspPrivateCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return c.DefaultLogger().WithField("Command", command)
}

func (c *LspPrivateCommand) noPermission() *message.PrivateMessage {
	return c.textReply("权限不够")
}

func (c *LspPrivateCommand) globalDisabledReply() *message.PrivateMessage {
	return c.textReply("无法操作该命令，该命令已被管理员禁用")
}

func (c *LspPrivateCommand) disabledReply() *message.PrivateMessage {
	return c.textSend("该命令已被设置为disable，请设置enable后重试")
}

func (c *LspPrivateCommand) notImplReply() *message.PrivateMessage {
	return c.textReply("暂未实现，你可以催作者GKD")
}

func (c *LspPrivateCommand) textSend(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return c.send(sendingMsg)
}

func (c *LspPrivateCommand) textReply(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	// reply work bad
	//sendingMsg.Append(message.NewPrivateReply(c.msg))
	sendingMsg.Append(message.NewText(text))
	return c.send(sendingMsg)
}

func (c *LspPrivateCommand) send(msg *message.SendingMessage) *message.PrivateMessage {
	return c.bot.SendPrivateMessage(c.uin(), msg)
}
func (c *LspPrivateCommand) sender() *message.Sender {
	return c.msg.Sender
}
func (c *LspPrivateCommand) uin() int64 {
	return c.sender().Uin
}

func (c *LspPrivateCommand) name() string {
	return c.sender().DisplayName()
}

func (c *LspPrivateCommand) NewMessageContext(log *logrus.Entry) *MessageContext {
	ctx := NewMessageContext()
	ctx.Source = SourceTypePrivate
	ctx.Lsp = c.l
	ctx.Log = log
	ctx.TextReply = func(text string) interface{} {
		return c.textReply(text)
	}
	ctx.Send = func(msg *message.SendingMessage) interface{} {
		return c.send(msg)
	}
	ctx.Reply = ctx.Send
	ctx.NoPermissionReply = func() interface{} {
		return c.noPermission()
	}
	ctx.DisabledReply = func() interface{} {
		ctx.Log.Debugf("disabled")
		return c.disabledReply()
	}
	ctx.GlobalDisabledReply = func() interface{} {
		ctx.Log.Debugf("global disabled")
		return c.globalDisabledReply()
	}
	ctx.Sender = c.sender()
	return ctx
}

func (c *LspPrivateCommand) checkGroupCode(groupCode int64) error {
	if groupCode == 0 {
		return fmt.Errorf("没有指定QQ群号码，请使用-g参数指定QQ群，例如对QQ群123456进行操作：%v %v %v", c.GetCmd(), "-g 123456", strings.Join(c.GetArgs(), " "))
	}
	group := c.bot.FindGroup(groupCode)
	if !c.l.PermissionStateManager.CheckRole(c.uin(), permission.Admin) {
		if group == nil {
			return errors.New("没有找到该QQ群，请确认bot是否在群内")
		}
		member := group.FindMember(c.uin())
		if member == nil {
			return errors.New("没有在该群内找到您，请确认您是否在群内")
		}
	} else {
		if group == nil {
			c.textReply("请注意未找到该群，如果bot刚刚启动，有可能是尚未刷新完毕，将继续查询数据")
		}
	}
	return nil
}
