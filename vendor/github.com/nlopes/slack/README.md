Slack API in Go [![GoDoc](https://godoc.org/github.com/nlopes/slack?status.svg)](https://godoc.org/github.com/nlopes/slack) [![Build Status](https://travis-ci.org/nlopes/slack.svg)](https://travis-ci.org/nlopes/slack)
===============

[![Join the chat at https://gitter.im/go-slack/Lobby](https://badges.gitter.im/go-slack/Lobby.svg)](https://gitter.im/go-slack/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

This library supports most if not all of the `api.slack.com` REST
calls, as well as the Real-Time Messaging protocol over websocket, in
a fully managed way.



## Change log
Support for the EventsAPI has recently been added. It is still in its early stages but nearly all events have been added and tested (except for those events in [Developer Preview](https://api.slack.com/slack-apps-preview) mode). API stability for events is not promised at this time.

### v0.2.0 - Feb 10, 2018

Release adds a bunch of functionality and improvements, mainly to give people a recent version to vendor against.

Please check [0.2.0](https://github.com/nlopes/slack/releases/tag/v0.2.0)

### CHANGELOG.md

 [CHANGELOG.md](https://github.com/nlopes/slack/blob/master/CHANGELOG.md) is available. Please visit it for updates.

## Installing

### *go get*

    $ go get -u github.com/nlopes/slack

## Example

### Getting all groups

```golang
import (
	"fmt"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("YOUR_TOKEN_HERE")
	// If you set debugging, it will log all requests to the console
	// Useful when encountering issues
	// api.SetDebug(true)
	groups, err := api.GetGroups(false)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
	}
}
```

### Getting User Information

```golang
import (
    "fmt"

    "github.com/nlopes/slack"
)

func main() {
    api := slack.New("YOUR_TOKEN_HERE")
    user, err := api.GetUserInfo("U023BECGF")
    if err != nil {
	    fmt.Printf("%s\n", err)
	    return
    }
    fmt.Printf("ID: %s, Fullname: %s, Email: %s\n", user.ID, user.Profile.RealName, user.Profile.Email)
}
```

## Minimal RTM usage:

See https://github.com/nlopes/slack/blob/master/examples/websocket/websocket.go


## Minimal EventsAPI usage:

See https://github.com/nlopes/slack/blob/master/examples/eventsapi/events.go


## Contributing

You are more than welcome to contribute to this project.  Fork and
make a Pull Request, or create an Issue if you see any problem.

## License

BSD 2 Clause license
