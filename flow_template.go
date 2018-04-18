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

	for _, key := range commandsConf.Keys() {
		generateCommands(&app.Commands, key, commandsConf.GetConfig(key))
	}

	app.RunAndExitOnError()

	return
}

func newAction(name string, conf config.Configuration) cli.ActionFunc {

	return func(ctx *cli.Context) (err error) {

		err = loadENV(ctx)

		if err != nil {
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

func generateCommands(cmds *[]cli.Command, name string, conf config.Configuration) {

	keys := conf.Keys()

	if len(keys) == 0 {
		return
	}

	objCount := 0
	for _, key := range keys {
		if conf.IsObject(key) || key == "usage" {
			objCount++
		}
	}

	// Command
	if objCount != len(keys) {
		*cmds = append(*cmds,
			cli.Command{
				Name:  name,
				Usage: conf.GetString("usage"),
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
				Action: newAction(name, conf),
			},
		)

		return
	}

	var subCommands []cli.Command

	for _, key := range conf.Keys() {

		if key == "usage" {
			continue
		}

		generateCommands(&subCommands, key, conf.GetConfig(key))

	}

	currentCommand := cli.Command{
		Name:        name,
		Usage:       conf.GetString("usage"),
		Subcommands: subCommands,
	}

	*cmds = append(*cmds, currentCommand)
}

`
)
