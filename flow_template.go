package main

var (
	flowTempl = `package main
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/gogap/logrus_mate"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	configStr = {{.config_str}}
)

func main() {

	conf := config.NewConfig(
		config.ConfigString(configStr),
	)

	appConf := conf.GetConfig("app")

	app := cli.NewApp()

	app.Version = appConf.GetString("version", "0.0.0")
	app.Author = appConf.GetString("author")
	app.Name = appConf.GetString("name", "app")
	app.Usage = appConf.GetString("usage")
	app.HelpName = app.Name

	if app.Version == "0.0.0" {
		app.HideVersion = true
	}

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
		configFiles := ctx.StringSlice("config")

		defaultConf := conf.GetConfig("default-config")

		mapConfigs := map[string]config.Configuration{
			"default-config": defaultConf,
		}

		for _, configArg := range configFiles {
			v := strings.SplitN(configArg, ":", 2)
			if len(v) == 1 {
				argsConfig := config.NewConfig(config.ConfigFile(v[0]))
				mapConfigs["default-config"] = argsConfig
			} else if len(v) == 2 {
				argsConfig := config.NewConfig(config.ConfigFile(v[1]))
				mapConfigs[v[0]] = argsConfig
			}
		}

		defaultConf = mapConfigs["default-config"]

		loggerConf := defaultConf.GetConfig("logger")

		logrus_mate.Hijack(
			logrus.StandardLogger(),
			logrus_mate.WithConfig(loggerConf),
		)

		flowCtx := context.NewContext()

		ctxList := ctx.StringSlice("ctx")
		ctxFiles := ctx.StringSlice("ctx-file")

		loadContext(ctxList, ctxFiles, flowCtx)

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

				if hConf, exist := mapConfigs[configName]; exist {
					handlerConf = hConf
				}

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

func loadContext(ctxList []string, ctxFiles []string, flowCtx context.Context) (err error) {

	if len(ctxList) == 0 && len(ctxFiles) == 0 {
		return
	}

	mapCtx := map[string]string{}

	for _, c := range ctxList {
		v := strings.SplitN(c, ":", 2)
		if len(v) != 2 {
			err = fmt.Errorf("ctx format error:%s", c)
			return
		}

		mapCtx[v[0]] = v[1]
	}

	for _, f := range ctxFiles {

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
			mapCtx[k] = v
		}
	}

	for k, v := range mapCtx {
		flowCtx.WithValue(k, v)
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
						Name:  "config, c",
						Usage: "use specified config to default config",
					},
					cli.StringSliceFlag{
						Name:  "env",
						Usage: "e.g.: --env USER:test --env PWD:asdf",
					},
					cli.StringSliceFlag{
						Name:  "ctx",
						Usage: "e.g.: --ctx code:gogap --env hello:world",
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
