package supermicro

//func NewFakeSupermicro() ironlib.Manager {
//
//	model := "SSG-6029P-E1CR12L-PH004"
//	var trace bool
//
//	// register inventory collectors
//	collectors := map[string]utils.Collector{
//		"ipmi":     utils.NewFakeIpmicfg(),
//		"smartctl": utils.NewSmartctlCmd(trace),
//		"storecli": utils.NewStoreCLICmd(trace),
//		"mlxup":    utils.NewMlxupCmd(trace),
//	}
//
//	uid, _ := uuid.NewRandom()
//	return &Supermicro{
//		ID:         uid.String(),
//		Vendor:     "supermicro",
//		Model:      utils.FormatProductName(model),
//		Dmidecode:  utils.NewFakeDmidecode(),
//		Collectors: collectors,
//		Logger:     logrus.New(),
//	}
//}
//
//func Test_ComponentUpdateAvailable(t *testing.T) {
//
//}
//
