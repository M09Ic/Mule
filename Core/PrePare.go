package Core

import (
	"Mule/utils"
	"context"
	"fmt"
)

func ScanPrepare(ctx context.Context, client *CustomClient, target string, root string) (map[string]*WildCard, error) {

	//defer utils.TimeCost()()
	//fmt.Println("start scan prepare")
	var WdMap map[string]*WildCard

	_, err := client.RunRequest(ctx, target, Additional{
		Mod:   "default",
		Value: "",
	})

	if err != nil {
		return nil, fmt.Errorf("cann't connect to %s\n", target)
	}

	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	//wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	WdMap, err = GenWildCardMap(ctx, client, RandomPath, target, root)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return WdMap, nil

}
