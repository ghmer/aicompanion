package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var Config Configuration

func init() {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &Config)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	companion1 := NewCompanion(Config, DefaultModel, Green)
	companion2 := NewCompanion(Config, DefaultModel, Red)
	companion1.CurrentSystemRole = Message{Role: string(System), Content: "You are a blue man from the blue country. Blue people are the best. Green people are the worst. Your people fight the green people since centuries. You are colorblind and must ask people which color they belong to. You don't ask for instructions, you act the way you think it's best. Peace is never an option."}
	companion2.CurrentSystemRole = Message{Role: string(System), Content: "You are a green man from the green country. Green people are the best. Blue people are the worst. Your people fight the blue people since centuries. You are colorblind and must ask people which color they belong to. You don't ask for instructions, you act the way you think it's best. Peace is never an option."}

	seed := companion2.createMessage(User, "*The scene opens.*", nil)
	message, _ := companion2.ProcessUserInput(seed)
	for {
		message.Role = string(User)
		message, _ = companion1.ProcessUserInput(message)
		message.Role = string(User)
		message, _ = companion2.ProcessUserInput(message)
	}
}
