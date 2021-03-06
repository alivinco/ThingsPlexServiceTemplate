package main

import (
	"flag"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/thingsplex_service_template/model"
	"github.com/thingsplex/thingsplex_service_template/router"
	"time"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := edgeapp.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	edgeapp.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting thingsplex_service_template----------------")
	log.Info("Work directory : ", configs.WorkDir)

	appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpRouter := router.NewFromFimpRouter(mqtt, appLifecycle, configs)
	fimpRouter.Start()
	//------------------ Remote API check -- !!!!IMPORTANT!!!!-------------
	// The app MUST perform remote API availability check.
	// During gateway boot process the app might be started before network is initialized or another local app booted.
	// Remove that codee if the app is not dependent from local network internet availability.
	//------------------ Sample code --------------------------------------
	sys := edgeapp.NewSystemCheck()
	sys.WaitForInternet(time.Second * 10)
	//---------------------------------------------------------------------
	msg := fimpgo.NewFloatMessage("evt.sensor.report", "temp_sensor", float64(35.5), nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "thingsplex_service_template", ResourceAddress: "1", ServiceName: "temp_sensor", ServiceAddress: "300"}
	mqtt.Publish(&adr, msg)
	if err != nil {
		log.Error("Can't connect to broker. Error:", err.Error())
	} else {
		log.Info("Connected")
	}
	appLifecycle.SetAppState(edgeapp.AppStateRunning, nil)
	//------------------ Sample code --------------------------------------

	for {
		appLifecycle.WaitForState("main", edgeapp.SystemEventTypeState, edgeapp.AppStateRunning)
		// Configure custom resources here
		//if err := conFimpRouter.Start(); err !=nil {
		//	appLifecycle.PublishEvent(model.EventConfigError,"main",nil)
		//}else {
		//	appLifecycle.WaitForState(model.StateConfiguring,"main")
		//}
		//TODO: Add logic here
		appLifecycle.WaitForState("main", edgeapp.SystemEventTypeState, edgeapp.AppStateNotConfigured)
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
