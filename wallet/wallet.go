package wallet

import "path/filepath"

const keystoreDirName = "keystore"
const MirasAccount = "0x77a13F4cf2cE723f0794F37eeC7635Dd65AE2736"
const AmiranAccount = "0x058DF0c85de392cc5bef6c749FE6DD8881a2CA44"
const BeknurAccount = "0x9Fd598035Ec0DD0909c054E855272b56F2BeC5C8"

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}
