package main

import (
	"DB_coursework/internal/admin_tool"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
)

func main() {
	fmt.Print("Добро пожаловать в приложение для добавления нового администратора\n")
	prompt := promptui.Prompt{
		Label:    "Введите имя пользователя базы данных",
		Validate: nil,
	}
	username, _ := prompt.Run()

	prompt.Label = "Введите пароль пользователя базы данных"
	prompt.Mask = '*'
	pass, _ := prompt.Run()

	tool, err := admin_tool.NewTool(os.Args[1], username, pass)
	if err != nil {
		fmt.Print("Во время инициализации приложения произошла ошибка\n", err)
		os.Exit(1)
	}
	fmt.Printf("Приветствую, %s!\n", username)
mainCycle:
	for {
		prompts := promptui.Select{
			Label: "Выберете действие",
			Items: []string{"Добавить администратора", "Удалить администратора", "Выйти"},
		}
		pos, _, _ := prompts.Run()
		if pos == 0 {
			var email, nickname, rights, password string
			prompt := promptui.Prompt{
				Label:    "Введите email администратора",
				Validate: nil,
			}
			for {
				email, _ = prompt.Run()
				exists, err := tool.CheckEmail(email)
				if err != nil {
					fmt.Print("Не удалось проверить уникальность email\n", err)
					continue mainCycle
				}
				if exists {
					prompts.Label = "Такой email уже существует, продолжить?"
					prompts.Items = []string{"Да", "Нет"}
					pos, _, _ = prompts.Run()
					if pos == 1 {
						continue mainCycle
					}
				} else {
					break
				}
			}

			prompt.Label = "Введите nickname администратора"
			for {
				nickname, _ = prompt.Run()
				exists, err := tool.CheckNickName(nickname)
				if err != nil {
					fmt.Print("Не удалось проверить уникальность nickname\n", err)
					continue mainCycle
				}
				if exists {
					prompts.Label = "Такой nickname уже существует, продолжить?"
					prompts.Items = []string{"Да", "Нет"}
					pos, _, _ = prompts.Run()
					if pos == 1 {
						continue mainCycle
					}
				} else {
					break
				}
			}

			prompts.Label = "Выберите права данного администратора"
			prompts.Items = []string{"Модерация пользователей", "Модерация контента", "Полная модерация"}
			pos, _, _ = prompts.Run()
			if pos == 0 {
				rights = "01"
			} else if pos == 1 {
				rights = "10"
			} else {
				rights = "11"
			}

			prompt.Label = "Введите пароль администратора"
			password, _ = prompt.Run()
			h2 := sha256.New()
			h2.Write([]byte(password))
			passHash := hex.EncodeToString(h2.Sum(nil))
			err = tool.AddAdmin(nickname, email, passHash, rights)
			if err != nil {
				fmt.Print("Не получается добавить администратора\n", err)
				continue mainCycle
			}
			fmt.Print("Администратор добавлен\n")
		} else if pos == 1 {
			prompt := promptui.Prompt{
				Label:    "Введите email или nickname администратора",
				Validate: nil,
			}
			tbd, _ := prompt.Run()
			err = tool.DelAdmin(tbd)
			if err != nil {
				fmt.Print("Не получается удалить администратора\n", err)
				continue mainCycle
			}
			fmt.Print("Администратор удален\n")
		} else {
			break
		}
	}
	tool.CloseConn()
}
