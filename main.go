package main

import (
	"fmt"
	"os"
)

var Config *Configuration
var AIScene *Scene

func init() {
	var err error
	Config, err = NewConfigFromFile("./config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	AIScene, err = NewSceneFromFile("./scene.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	companion1 := NewCompanion(*Config, DefaultModel, Green)
	companion2 := NewCompanion(*Config, DefaultModel, Red)
	companion1.CurrentSystemRole = Message{Role: string(System), Content: AIScene.Assistant1}
	companion2.CurrentSystemRole = Message{Role: string(System), Content: AIScene.Assistant2}

	message := companion1.createMessage(User, AIScene.OpeningMessage, nil)

	var assistant1 bool = true

	for {
		if assistant1 {
			message, _ = companion1.ProcessUserInput(message)
		} else {
			message, _ = companion2.ProcessUserInput(message)
		}
		message.Role = string(User)
		assistant1 = !assistant1
	}
}
