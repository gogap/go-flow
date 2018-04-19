package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/gogap/builder"
	"github.com/gogap/config"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "go-flow"
	app.HelpName = "go-flow"
	app.HideVersion = true

	app.Commands = cli.Commands{
		cli.Command{
			Name:  "build",
			Usage: "build your own flow into binary",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Usage: "flow config file",
				},
			},
			Action: build,
		},

		cli.Command{
			Name:  "run",
			Usage: "run flow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Usage: "flow config file",
				},
			},
			Action:          run,
			SkipFlagParsing: true,
			SkipArgReorder:  true,
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "be verbose",
		},
	}

	app.RunAndExitOnError()
}

func createBuilder(appName string, verboseOfGoBuild, verboseOfGoGet bool, conf config.Configuration) (bu *builder.Builder, err error) {

	argsConf := ""

	if verboseOfGoGet {
		argsConf += "go-get = [\"-v\"]\n"
	}

	if verboseOfGoBuild {
		argsConf += "go-build = [\"-v\"]\n"
	}

	buildConfStr := fmt.Sprintf(`%s {
packages = %s
build.args {
    %s
  }
}`, appName, conf.GetStringList("packages"), argsConf)

	tmpl, err := template.New(appName).Parse(flowTempl)
	if err != nil {
		return
	}

	b, err := builder.NewBuilder(
		builder.ConfigString(buildConfStr),
		builder.Template(tmpl),
	)

	if err != nil {
		return
	}

	bu = b
	return
}

func build(ctx *cli.Context) (err error) {

	configFile := ctx.String("config")

	if len(configFile) == 0 {
		err = fmt.Errorf("please input config file")
		return
	}

	conf := config.NewConfig(config.ConfigFile(configFile))

	appName := conf.GetString("app.name", "app")

	verbose := ctx.Parent().Bool("verbose")

	b, err := createBuilder(appName, verbose, verbose, conf)
	if err != nil {
		return
	}

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = b.Build(map[string]interface{}{"config_str": fmt.Sprintf("`%s`", string(configData))}, appName)

	return
}

func run(ctx *cli.Context) (err error) {

	set := flag.NewFlagSet("run", 0)

	confArg := set.String("config", "", "flow config file")

	err = set.Parse(ctx.Args()[0:2])
	if err != nil {
		return
	}

	configFile := *confArg

	if len(configFile) == 0 {
		err = fmt.Errorf("please input config file")
		return
	}

	conf := config.NewConfig(config.ConfigFile(configFile))

	appName := conf.GetString("app.name", "app")

	verbose := ctx.Parent().Bool("verbose")

	b, err := createBuilder(appName, false, verbose, conf)
	if err != nil {
		return
	}

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	err = b.Run(map[string]interface{}{"config_str": fmt.Sprintf("`%s`", string(configData))}, appName, ctx.Args()[2:])

	return
}
