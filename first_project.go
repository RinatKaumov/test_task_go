package main

import (
    "fmt"
    "log"
    "time"
    "strconv"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для хранения информации о лекарствах
type Medicine struct {
    Name     string
    Expiry   time.Time
}

var medicineCabinet = make(map[string]Medicine)

func main() {
    // Подключаемся к боту через токен
    bot, err := tgbotapi.NewBotAPI("7483565697:AAGAYz-LpXI0uiJPmfYKIzcd2rZwnoWwQKY")
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil { // Игнорируем пустые сообщения
            continue
        }

        switch update.Message.Command() {
        case "start":
            msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я помогу тебе вести учет лекарств. Используй команды:\n/add <название> <срок в формате ДД-ММ-ГГГГ> - добавить лекарство\n/list - показать все лекарства\n/delete <название> - удалить лекарство\n/expiring - лекарства с истекающим сроком")
            bot.Send(msg)

        case "add":
            args := update.Message.CommandArguments()
            if args == "" {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите название лекарства и срок годности в формате: /add <название> <срок ДД-ММ-ГГГГ>"))
                continue
            }

            // Разделяем аргументы на название и срок
            parts := strings.SplitN(args, " ", 2)
            if len(parts) < 2 {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный формат. Используй: /add <название> <срок ДД-ММ-ГГГГ>"))
                continue
            }

            name := parts[0]
            expiryDate, err := time.Parse("02-01-2006", parts[1])
            if err != nil {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат даты. Используй ДД-ММ-ГГГГ."))
                continue
            }

            medicineCabinet[name] = Medicine{Name: name, Expiry: expiryDate}
            bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Лекарство '%s' добавлено со сроком годности до %s.", name, expiryDate.Format("02-01-2006"))))

        case "list":
            if len(medicineCabinet) == 0 {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "В твоей аптечке пока нет лекарств."))
            } else {
                var message strings.Builder
                message.WriteString("Текущие лекарства в аптечке:\n")
                for name, med := range medicineCabinet {
                    message.WriteString(fmt.Sprintf("%s - срок годности до %s\n", name, med.Expiry.Format("02-01-2006")))
                }
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, message.String()))
            }

        case "delete":
            name := update.Message.CommandArguments()
            if _, exists := medicineCabinet[name]; exists {
                delete(medicineCabinet, name)
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Лекарство '%s' удалено из аптечки.", name)))
            } else {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Лекарство '%s' не найдено.", name)))
            }

        case "expiring":
            var expiringMeds []string
            today := time.Now()

            for name, med := range medicineCabinet {
                if med.Expiry.Before(today.AddDate(0, 1, 0)) { // Лекарства, срок которых истекает через месяц
                    expiringMeds = append(expiringMeds, fmt.Sprintf("%s - до %s", name, med.Expiry.Format("02-01-2006")))
                }
            }

            if len(expiringMeds) == 0 {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нет лекарств, срок годности которых истекает в ближайший месяц."))
            } else {
                bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Лекарства, срок годности которых скоро истекает:\n%s", strings.Join(expiringMeds, "\n"))))
            }

        default:
            msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не понимаю эту команду. Попробуй /start, чтобы увидеть список доступных команд.")
            bot.Send(msg)
        }
    }
}
