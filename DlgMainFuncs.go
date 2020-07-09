package main

func (me *DlgMain) addFilesIfNotYet(mp3s []string) {
	me.lstFiles.SetRedraw(false)
	for _, mp3 := range mp3s {
		if me.lstFiles.FindItem(mp3) == nil { // not yet in the list
			me.lstFiles.AddItemWithIcon(mp3, 0) // will fire LVN_INSERTITEM
		}
	}
	me.lstFiles.SetRedraw(true)
	me.lstFiles.Column(0).FillRoom()
}
