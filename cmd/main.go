package main

import (
	"StakemeBotV2/internal/app"
)

func main() {
	app := app.NewApp()
	app.Init()
	app.StartWsCon()
	app.StartTgBot()
	app.StartMntr()
	app.Wait()

	// for {	
	// 	for _, chain := range cnfg.Chains {
	// 		count, err :=  db.CountSignedBlocks(dbcon, chain.Name); 
	// 		if err != nil {
	// 			logger.WithFields(logrus.Fields{
	// 				"module": "db",
	// 				"chain": chain.Name,
	// 			}).Errorf("Get block count for chain %s failed %s", chain.Name, err)
	// 		}
	// 		logger.WithFields(logrus.Fields{
	// 			"module": "db",
	// 			"chain": chain.Name,
	// 		}).Infof("Block count for chain %s: %d", chain.Name, count)
	// 	}
		
	// 	logger.WithFields(logrus.Fields{
	// 		"module": "app",
	// 		"cpu":  runtime.NumCPU(),
	// 		"goroutines": runtime.NumGoroutine(),
	// 	}).Warnf("Application load")
	// 	time.Sleep(5 * time.Second)
	// }	
}


