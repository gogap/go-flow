package main

var (
	flowTempl = `package main
import (
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/urfave/cli"
)

var (
	configStr = {{.config_str}}
)

func main() {

	conf := config.NewConfig(
		config.ConfigString(configStr),
	)

	app := cli.NewApp()
	app.HideVersion = true

	appConf := conf.GetConfig("app")

	app.Name = appConf.GetString("name", "app")
	app.Usage = appConf.GetString("usage")
	app.HelpName = app.Name

	commandsConf := appConf.GetConfig("commands")

	for _, cmdName := range commandsConf.Keys() {
		cmdConf := commandsConf.GetConfig(cmdName)
		app.Commands = append(app.Commands,
			cli.Command{
				Name:  cmdName,
				Usage: cmdConf.GetString("usage"),
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "disable, d",
						Usage: "disable steps, e.g.: -d devops.aliyun.cs.cluster.deleted.wait -d devops.aliyun.cs.cluster.running.wait",
					},
					cli.StringSliceFlag{
						Name:  "context, ctx",
						Usage: "flow context, e.g.: -ctx a:b -ctx c:d",
					},
				},
				Action: newAction(cmdName, cmdConf),
			},
		)
	}

	app.RunAndExitOnError()

	return
}

func newAction(name string, conf config.Configuration) cli.ActionFunc {

	return func(ctx *cli.Context) error {

		disableSteps := ctx.StringSlice("disable")

		defaultConf := conf.GetConfig("default-config")

		flowCtx := context.NewContext()

		trans := flow.Begin(flowCtx, config.WithConfig(defaultConf))

		flowList := conf.GetStringList("flow")

		flowItemConfig := conf.GetConfig("config")

		mapDisabelSteps := map[string]bool{}

		for _, step := range disableSteps {
			mapDisabelSteps[step] = true
		}

		for _, item := range flowList {

			handlerAndConf := strings.SplitN(item, "@", 2)

			name := ""
			configName := ""

			if len(handlerAndConf) == 2 {
				name = handlerAndConf[0]
				configName = handlerAndConf[1]
			} else {
				name = handlerAndConf[0]
			}

			if mapDisabelSteps[name] {
				continue
			}

			if len(configName) > 0 {
				handlerConf := flowItemConfig.GetConfig(configName)
				trans.Then(name, config.WithConfig(handlerConf))
			} else {
				trans.Then(name)
			}
		}

		return trans.Commit()
	}
}`
)
