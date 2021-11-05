package main

import "mighty/cmd"

func main() {
	cmd.Execute()

	//
	//apiClient, err := api.New(, "")
	//if err != nil {
	//	panic(err)
	//}
	//allHistoricEntries, err := apiClient.FetchEntries("30w")
	//xl := export.ExcelFile("mite-entries.xlsx")
	//err = xl.SaveAllEntries(allHistoricEntries); if err != nil {
	//	fmt.Println(err)
	//}
	//
	//duration, err := time.ParseDuration("1m")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("......")
	//time.Sleep(duration)
	//fmt.Println("......")
	//
	//err = xl.ReloadFromDisk()
	//if err != nil {
	//		fmt.Println(err)
	//}
	//
	//xl.GetSheets()
	//entries := xl.ReadAllEntriesBySheet("November 2021")
	//fmt.Printf("%v",entries)
	//err = apiClient.SendEntriesToMite(entries)
	//if err != nil {
	//	fmt.Println(err)
	//}

}
