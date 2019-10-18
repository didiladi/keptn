package handlers

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/google/uuid"

	datastore "github.com/keptn/go-utils/pkg/mongodb-datastore/utils"
	keptnutils "github.com/keptn/go-utils/pkg/utils"

	"github.com/keptn/keptn/api/models"
	"github.com/keptn/keptn/api/restapi/operations/event"
	"github.com/keptn/keptn/api/utils"
	"github.com/keptn/keptn/api/ws"
)

type eventData struct {
	models.Data  `json:",inline"`
	EventContext models.EventContext `json:"eventContext"`
}

// PostEventHandlerFunc forwards a received event to the event broker
func PostEventHandlerFunc(params event.SendEventParams, principal *models.Principal) middleware.Responder {

	keptnContext := uuid.New().String()
	l := keptnutils.NewLogger(keptnContext, "", "api")
	l.Info("API received a keptn event")

	token, err := ws.CreateChannelInfo(keptnContext)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating channel info %s", err.Error()))
		return sendInternalError(err)
	}

	source, _ := url.Parse("https://github.com/keptn/keptn/api")
	contentType := "application/json"
	eventContext := models.EventContext{KeptnContext: &keptnContext, Token: &token}
	forwardData := eventData{Data: params.Body.Data, EventContext: eventContext}

	ev := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        string(params.Body.Type),
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": keptnContext},
		}.AsV02(),
		Data: forwardData,
	}

	_, _, err = utils.PostToEventBroker(ev)
	if err != nil {
		l.Error(fmt.Sprintf("Error sending CloudEvent %s", err.Error()))
		return sendInternalError(err)
	}

	return event.NewSendEventOK().WithPayload(&eventContext)
}

// GetEventHandlerFunc returns an event specified by keptnContext and eventType
func GetEventHandlerFunc(params event.GetEventParams, principal *models.Principal) middleware.Responder {
	eventHandler := datastore.NewEventHandler(getDatastoreURL())

	cloudEvent, err := eventHandler.GetEvent(*params.KeptnContext, *params.Type)
	if err != nil {
		return sendInternalError(fmt.Errorf("%s", err.Message))
	}

	fmt.Print(cloudEvent.Shkeptncontext)

	//response := models.Event{}

	return event.NewSendEventOK().WithPayload(nil)
}

func sendInternalError(err error) *event.SendEventDefault {
	return event.NewSendEventDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
}

func getDatastoreURL() string {
	return "http://" + os.Getenv("DATASTORE_URI")
}
