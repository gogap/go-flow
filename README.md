go-flow
=======

Run flow without any code!


### Install

**Requirement:**

- go
- git


```bash
go get -v github.com/gogap/go-flow
```


### Run your first flow


The config file of `hello.conf`

```hocon
packages = ["github.com/flow-contrib/example/hello"]

app {
    name = "hello"
    usage = "This is a demo for run flow"

    commands {
        say {
            usage = "This command will print hello"

            default-config = { name = "gogap" }

            flow = ["example.hello", "example.hello@confA"]
            
            config = {
              confA = {
                 name = "Zeal"
            }
        }
    }
}
```

```bash
$ go-flow -v run --config hello.conf say

Hello: gogap
Hello: Zeal
```

### Distribution your flow

```bash
$ go-flow build --config hello.conf
```

or 

```bash
$ GOOS=linux GOARCH=amd64 go-flow build --config hello.conf
```

Just send the output binary `./hello` to your firend, then He just input command:

```
$ ./hello say

Hello: gogap
Hello: Zeal
```


### Advance Example

#### Subcommands

```
packages = ["github.com/flow-contrib/example/hello"]

app {
    name = "hello"
    usage = "This is a demo for run flow"

    commands {
        say {
            usage = "say something"

            gogap {
                usage = "This command will print hello"

                default-config = { name = "gogap" }

                flow = ["example.hello", "example.hello@confA"]
                
                config = {
                  confA = {
                     name = "Zeal"
                    }
                }
            }

            world {
                usage = "This command will print world"

                default-config = { name = "world" }

                flow = ["example.hello"]
                
                config = {}
            }
        }
    }
}
```

```bash
$ go-flow run --config hello.conf say
NAME:
   hello say - say something

USAGE:
   hello say command [command options] [arguments...]

COMMANDS:
     gogap  This command will print hello
     world  This command will print world
```

```bash
$ go-flow run --config hello.conf say gogap
Hello: gogap
Hello: Zeal
```

```bash
$ go-flow run --config hello.conf say world
Hello: world
```


#### Run javascript flow

`goja.conf`

```hocon
packages = ["github.com/flow-contrib/goja"]

app {
	name = "goja"
	usage = ""

	commands {
		execute {
			usage = "execute javascript"

			default-config = {}

			flow = ["lang.javascript.goja@confA", "lang.javascript.goja@confB"]

			config {
				confA = {src = A.js}
				confB = {src = B.js}
			}
		}
	}
}
```


`A.js`

```javascript
console.log("I am from A.js")
```

`B.js`

```javascript
console.log("I am from B.js")
```

```bash
$ go-flow run execute --config goja.conf 

2018/04/17 22:59:24 A.js
2018/04/17 22:59:24 B.js
```

