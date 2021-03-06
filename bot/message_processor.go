package bot

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

const panopticon = "'The [tockers] must never know whether [they are] looked at at any one moment; but [they] must be sure that [they] may always be so' - Foucault, Discipline 201"

// processMessage handles incomming messages
func (bot *Bot) processMessage(message *slack.MessageEvent) {
	user := message.User
	botID := bot.Slack.GetSelfID()
	// Handle Violators
	userID := bot.violatorUserMap.Get(user)
	if userID != "" {
		bot.violatorMessage(message, user)
	}

	botCalled := strings.HasPrefix(
		message.Text,
		fmt.Sprintf("<@%s>", botID),
	)
	if botCalled { // Messages made directly to bot
		switch {
		case bot.isMasterUser(user):
			{
				bot.masterMessages(message)
			}
		default:
			{
				bot.niceMessage(message, user)
			}
		}
	} else {
		switch {
		// Messages that contain the word tick
		case strings.Contains(message.Text, " tick "):
			{
				bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(
					"tock", message.Channel,
				))
			}
		// Messages that references the bot will be send out 30% of the time. See: Foucault, Discipline 201
		case strings.Contains(message.Text, fmt.Sprintf("<@%s>", botID)):
			{
				var returnMessage string
				randomInt := rand.Intn(100)
				if randomInt >= 70 {
					returnMessage = bot.MessageRepo.Nice.GenerateMessage(user)
				} else if randomInt <= 3 {
					returnMessage = panopticon
				}
				bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(
					returnMessage,
					message.Channel,
				))
			}
		}
	}
}

// violatorMessage has the message for a late user
func (bot *Bot) violatorMessage(message *slack.MessageEvent, user string) {
	var returnMessage string
	// Check if user is still late
	if bot.isLateUser(user) {
		returnMessage = bot.MessageRepo.Angry.GenerateMessage(user)
	} else {
		returnMessage = fmt.Sprintf(
			"<@%s>, I was about to yell at you, but then I realized you actually filled out your timesheet. Thanks! ^_^",
			user,
		)
	}
	bot.violatorUserMap.Delete(user)
	bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(returnMessage, message.Channel))
}

// masterMessages contains the commands for admins
func (bot *Bot) masterMessages(message *slack.MessageEvent) {
	var returnMessage string
	botID := bot.Slack.GetSelfID()
	switch {
	case strings.Contains(message.Text, "slap users"):
		{
			go bot.SlapLateUsers()
			returnMessage = "Slapping Users!"
		}
	case strings.Contains(message.Text, "remind users"):
		{
			braketFinder := regexp.MustCompile("{{.*?}}")
			foundMessages := braketFinder.FindAllString(message.Text, 1)
			if len(foundMessages) == 0 {
				returnMessage = "Error: no message to send or message not formatted correctly"
			} else {
				messageToSend := strings.Trim(foundMessages[0], "{}")
				go bot.RemindUsers(messageToSend)
				returnMessage = fmt.Sprintf("Reminding users with `%s`", messageToSend)
			}
		}
	case strings.Contains(message.Text, "bother users"):
		{
			bot.startviolatorUserMapUpdater()
			returnMessage = "Starting to bother users!"
		}
	case strings.Contains(message.Text, "who is late?"):
		{
			lateList, total := bot.fetchLateUsers()
			returnMessage = fmt.Sprintf("%s are late! %d people total.", lateList, total)
		}
	default:
		{
			returnMessage = fmt.Sprintf(
				"Commands:\n Message tardy users `<@%s>: slap users!`\n Remind users nicely `<@%s>: remind users {{Text of message here}}`\nBother tardy users `<@%s>: bother users!`\nFind out who is late `<@%s>: who is late?`",
				botID,
				botID,
				botID,
				botID,
			)
		}
	}
	bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(
		returnMessage, message.Channel,
	))
}

// niceMessage are commands for user who are not late
func (bot *Bot) niceMessage(message *slack.MessageEvent, user string) {
	var returnMessage string
	switch {
	case strings.Contains(message.Text, "hello"):
		{
			bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(
				bot.MessageRepo.Nice.GenerateMessage(user),
				message.Channel,
			))
		}
	case strings.Contains(message.Text, "status"):
		{
			go func() {
				if bot.isLateUser(user) {
					returnMessage = fmt.Sprintf("<@%s>, you're late -_-", user)
				} else {
					returnMessage = fmt.Sprintf("<@%s>, you're on time! ^_^", user)
				}
				bot.Slack.SendMessage(bot.Slack.NewOutgoingMessage(
					returnMessage, message.Channel,
				))
			}()
		}
	}
}
