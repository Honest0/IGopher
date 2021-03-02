package igopher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asticode/go-astilectron"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type SucessState string

const (
	SUCCESS SucessState = "Success"
	ERROR   SucessState = "Error"
)

var (
	config                          BotConfigYaml
	validate                        = validator.New()
	reloadCh, hotReloadCh, exitedCh chan bool
	ctx                             context.Context
	cancel                          context.CancelFunc
)

// MessageOut represents a message for electron (going out)
type MessageOut struct {
	Status  SucessState `json:"status"`
	Msg     string      `json:"msg"`
	Payload interface{} `json:"payload,omitempty"`
}

// MessageIn represents a message from electron (going in)
type MessageIn struct {
	Msg     string          `json:"msg"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// This will send a message to Electron Gui and execute a callback
// Callback function is optional
func sendMessageToElectron(msg MessageOut, callback func(m *astilectron.EventMessage)) {
	w.SendMessage(msg, callback)
}

// Handling function for incoming messages
func handleMessages() {
	w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		// Unmarshal
		var i MessageIn
		var err error
		if err = m.Unmarshal(&i); err != nil {
			logrus.Errorf("Unmarshaling message %+v failed: %v", *m, err)
			return MessageOut{Status: "Error during message reception"}
		}

		// Process message
		config = ImportConfig()
		switch i.Msg {
		case "resetGlobalDefaultSettings":
			return i.resetGlobalSettingsCallback()

		case "clearAllData":
			return i.clearDataCallback()

		case "igCredentialsForm":
			return i.credentialsFormCallback()

		case "quotasForm":
			return i.quotasFormCallback()

		case "schedulerForm":
			return i.schedulerCallback()

		case "blacklistForm":
			return i.blacklistFormCallback()

		case "dmSettingsForm":
			return i.dmBotFormCallback()

		case "dmUserScrappingSettingsForm":
			return i.dmScrapperFormCallback()

		case "launchDmBot":
			return i.launchDmBotCallback()

		case "stopDmBot":
			return i.stopDmBotCallback()

		case "hotReloadBot":
			return i.hotReloadCallback()

		case "getLogs":
			return i.getLogsCallback()

		default:
			logrus.Error("Unexpected message received.")
			return MessageOut{Status: ERROR}
		}
	})
}

/* Callback functiosn to handle electron messages */

func (m *MessageIn) resetGlobalSettingsCallback() MessageOut {
	config = ResetBotConfig()
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Global configuration was successfully reseted!"}
}

func (m *MessageIn) clearDataCallback() MessageOut {
	if err := ClearData(); err != nil {
		return MessageOut{Status: ERROR, Msg: fmt.Sprintf("IGopher data clearing failed! Error: %v", err)}
	}
	return MessageOut{Status: SUCCESS, Msg: "IGopher data successfully cleared!"}
}

func (m *MessageIn) credentialsFormCallback() MessageOut {
	var err error
	var credentialsConfig AccountYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &credentialsConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(credentialsConfig)
	if err != nil {
		logrus.Warning("Validation issue on credentials form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on credentials form, please check given informations."}
	}

	config.Account = credentialsConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Credentials settings successfully updated!"}
}

func (m *MessageIn) quotasFormCallback() MessageOut {
	var err error
	var quotasConfig QuotasYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &quotasConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(quotasConfig)
	if err != nil {
		logrus.Warning("Validation issue on quotas form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on quotas form, please check given informations."}
	}

	config.Quotas = quotasConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Quotas settings successfully updated!"}
}

func (m *MessageIn) schedulerCallback() MessageOut {
	var err error
	var schedulerConfig ScheduleYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &schedulerConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(schedulerConfig)
	if err != nil {
		logrus.Warning("Validation issue on scheduler form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on scheduler form, please check given informations."}
	}

	config.Schedule = schedulerConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Scheduler settings successfully updated!"}
}

func (m *MessageIn) blacklistFormCallback() MessageOut {
	var err error
	var blacklistConfig BlacklistYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &blacklistConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(blacklistConfig)
	if err != nil {
		logrus.Warning("Validation issue on blacklist form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on blacklist form, please check given informations."}
	}

	config.Blacklist = blacklistConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Blacklist settings successfully updated!"}
}

func (m *MessageIn) dmBotFormCallback() MessageOut {
	var err error
	var dmConfig AutoDmYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &dmConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(dmConfig)
	if err != nil {
		logrus.Warning("Validation issue on dm tool form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on dm tool form, please check given informations."}
	}

	config.AutoDm = dmConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Dm bot settings successfully updated!"}
}

func (m *MessageIn) dmScrapperFormCallback() MessageOut {
	var err error
	var scrapperConfig ScrapperYaml
	// Unmarshal payload
	if err = json.Unmarshal([]byte(m.Payload), &scrapperConfig); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return MessageOut{Status: ERROR, Msg: "Failed to unmarshal message payload."}
	}

	err = validate.Struct(scrapperConfig)
	if err != nil {
		logrus.Warning("Validation issue on scrapper form, abort.")
		return MessageOut{Status: ERROR, Msg: "Validation issue on scrapper form, please check given informations."}
	}

	config.SrcUsers = scrapperConfig
	ExportConfig(config)
	return MessageOut{Status: SUCCESS, Msg: "Scrapper settings successfully updated!"}
}

func (m *MessageIn) launchDmBotCallback() MessageOut {
	var err error
	if err = CheckConfigValidity(); err == nil {
		ctx, cancel = context.WithCancel(context.Background())
		go launchDmBot(ctx)
		return MessageOut{Status: SUCCESS, Msg: "Dm bot successfully launched!"}
	}
	return MessageOut{Status: ERROR, Msg: err.Error()}
}

func (m *MessageIn) stopDmBotCallback() MessageOut {
	if exitedCh != nil {
		cancel()
		res := <-exitedCh
		if res {
			return MessageOut{Status: SUCCESS, Msg: "Dm bot successfully stopped!"}
		}
		return MessageOut{Status: ERROR, Msg: "Error during bot stopping! Please restart IGopher"}
	}
	return MessageOut{Status: ERROR, Msg: "Bot is in the initialization phase, please wait before trying to stop it."}
}

func (m *MessageIn) hotReloadCallback() MessageOut {
	if BotStruct.running {
		if hotReloadCh != nil {
			hotReloadCh <- true
			res := <-hotReloadCh
			if res {
				return MessageOut{Status: SUCCESS, Msg: "Bot hot reload successfully!"}
			}
			return MessageOut{Status: ERROR, Msg: "Error during bot hot reload! Please restart the bot"}
		}
		return MessageOut{Status: ERROR, Msg: "Bot is in the initialization phase, please wait before trying to hot reload it."}
	}
	return MessageOut{Status: ERROR, Msg: "Bot isn't running yet."}
}

func (m *MessageIn) getLogsCallback() MessageOut {
	logs, err := parseLogsToString()
	if err != nil {
		logrus.Errorf("Can't parse logs: %v", err)
		return MessageOut{Status: ERROR, Msg: fmt.Sprintf("Can't parse logs: %v", err)}
	}
	logrus.Debug("Logs fetched successfully!")
	return MessageOut{Status: SUCCESS, Msg: logs}
}