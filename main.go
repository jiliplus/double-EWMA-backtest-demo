package main

func main() {
	srcName := "./btcusdt.sqlite3"
	//
	db := openToMemory(srcName)
	defer db.Close()
	//
	tickSrc(db, nil)
}
