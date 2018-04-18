package main

var (
	flowTempl = `package main
import (
	"strings"
	"fmt"
	"bytes"
	"os"
	"encoding/json"
	"io/ioutil"

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
						Name:  "env",
						Usage: "e.g.: --env USER:test --env PWD:asdf",
					},
					cli.StringSliceFlag{
						Name:  "env-file",
						Usage: "e.g.: --env-file a.json --env-file b.json",
					},
				},
				Action: newAction(cmdName, cmdConf),
			},
		)
	}

	app.Commands = append(app.Commands,
		cli.Command{
			Name:  "metadata",
			Usage: "show this flow metadata",
			Subcommands: cli.Commands{
				cli.Command{
					Name:  "config",
					Usage: "build and run config",
					Action: func(ctx *cli.Context) error {
						fmt.Println(configStr)
						return nil
					},
				},
			},
		},
	)

	app.RunAndExitOnError()

	return
}

func newAction(name string, conf config.Configuration) cli.ActionFunc {

	return func(ctx *cli.Context) (err error) {

		err = loadENV(ctx)

		if err!=nil {
			return
		}

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

			if mapDisabelSteps[item] {
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
}

func loadENV(ctx *cli.Context) (err error) {
	envs := ctx.StringSlice("env")
	envFiles := ctx.StringSlice("env-file")

	if len(envs) == 0 && len(envFiles) == 0 {
		return
	}

	mapENV := map[string]string{}

	for _, env := range envs {
		v := strings.SplitN(env, ":", 2)
		if len(v) != 2 {
			err = fmt.Errorf("env format error:%s", env)
			return
		}

		mapENV[v[0]] = v[1]
	}

	for _, f := range envFiles {

		var data []byte
		data, err = ioutil.ReadFile(f)
		if err != nil {
			return
		}

		buf := bytes.NewBuffer(data)
		decoder := json.NewDecoder(buf)
		decoder.UseNumber()

		tmpMap := map[string]string{}
		err = decoder.Decode(&tmpMap)
		if err != nil {
			return
		}

		for k, v := range tmpMap {
			mapENV[k] = v
		}
	}

	for k, v := range mapENV {
		os.Setenv(k, v)
	}

	return
}
`
)
