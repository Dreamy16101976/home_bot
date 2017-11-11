/*  home_bot - Telegram bot for Smart Home
    Copyright (C) 2017 - Alexey "FoxyLab" Voronin
    Email:    support@foxylab.com
    Website:  https://acdc.foxylab.com

    This program is free software; you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation; either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program; if not, write to the Free Software
    Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307 USA

    CREDITS
    Brian C. Lane - digitemp https://www.digitemp.com/
    Ted Burke - RobotEyez https://github.com/tedburke/RobotEyez
*/ 

package main

import (
	"flag"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"github.com/Syfaro/telegram-bot-api" ////go get github.com/Syfaro/telegram-bot-api
	"golang.org/x/image/bmp" //go get golang.org/x/image/bmp
)

var (
	botToken   string //токен бота
	godString  string //ID бога (строка)
	god        int //ID бога (число)
	bmpFile    string = "frame.bmp" //имя BMP-файла
	jpgFile    string = "frame.jpg" //имя JPEG-файла
	jpgQuality int    = 80          //качество JPEG-файла
	spyCmd     string = "\\spy.cmd" //командный файл для получения снимка
	tempCmd    string = "\\temp.cmd" //командный файл для считывания температуры
)

func init() {
	flag.StringVar(&botToken, "bot", "", "Bot Token") //-bot
	flag.StringVar(&godString, "god", "", "God ID")   //-god
	flag.Parse()
	//требуется задание токена
	if botToken == "" {
		log.Print("-bot is required")
		os.Exit(1)
	}
	//трбуется задание ID бога
	if godString == "" {
		log.Print("-god is required")
		os.Exit(1)
	}
}

func main() {
	//получение ID бога
	god64, err := strconv.ParseInt(godString, 10, 0)
	if err != nil {
		log.Panic(err)
	}
	god = int(god64)
	//создание экземляра бота
	bot, err := tgbotapi.NewBotAPI(botToken) 
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	//создание структуры для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u) //создание канала
	//обработка обновлений
	for update := range updates {
		// универсальный ответ на любое сообщение
		reply := "???"
		if update.Message.From.ID != god { //игнорирование сообщения, если оно не от бога
			continue
		}
		if update.Message == nil { //игнорирование пустого сообщения
			continue
		}
		//протоколирование сообщения
		log.Printf("[%s] %d %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)
		//обработка команд
		switch update.Message.Command() {
		case "start":
			reply = "Привет, " + update.Message.From.UserName + "\n Команды: \n /start - приветствие \n /help - справка \n /temp - температура \n /cam - снимок \n /about - о боте"
			//создание ответного сообщения
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			//отправка ответного сообщения
			bot.Send(msg)
		case "help":
			reply = "Команды: \n /start - приветствие \n /help - справка \n /temp - температура \n /cam - снимок \n /about - о боте"
			//создание ответного сообщения
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			//отправка ответного сообщения
			bot.Send(msg)
		case "about":
			reply = "Бот \"Умного дома\" \n /help - список команд"
			//создание ответного сообщения
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			//отправка ответного сообщения
			bot.Send(msg)
		case "temp":
			//получение текущей директории
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			os.Chdir(pwd)
			cmdf := pwd + tempCmd //формирование команды для получения температуры
			//выполнение команды
			out, err := exec.Command(cmdf, "").Output()
			if err != nil {
				log.Fatal(err)
			}
			//анализ ответа
			outs := strings.Split(string(out), "\r\n")
			//формирование сообщения о температуре
			if len(outs) > 1 {
				reply = "Температура " + outs[2] + " °C"
			} else {
				reply = "Температура ---" + " °C"
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			//отправка сообщения о температуре
			bot.Send(msg)
		case "cam":
			//получение текущей директории
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			os.Chdir(pwd)
			cmdf := pwd + spyCmd //формирование команды для получения снимка
			cmd := exec.Command(cmdf, "") //выполнение команды
			cmd.Run()
			//преобразование BMP в JPEG
			imgfile, err := os.Open(bmpFile)
			if err != nil {
				panic(err)
			}
			img, err := bmp.Decode(imgfile)
			imgfile.Close()
			out, err := os.Create(jpgFile)
			if err != nil {
				panic(err)
			}
			var opt jpeg.Options
			opt.Quality = jpgQuality
			jpeg.Encode(out, img, &opt)
			out.Close()
			//формирование файла для отправки
			f, err := os.Open(jpgFile)
			if err != nil {
				panic(err)
			}
			reader := tgbotapi.FileReader{Name: jpgFile, Reader: f, Size: -1}
			file := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, reader)
			bot.Send(file) //отправка файла со снимком
			f.Close()
		}
	}
}
