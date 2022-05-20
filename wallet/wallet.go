package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"path/filepath"
)

const keystoreDirName = "keystore"
const MirasAccount = "0x77a13F4cf2cE723f0794F37eeC7635Dd65AE2736"
const AmiranAccount = "0x058DF0c85de392cc5bef6c749FE6DD8881a2CA44"
const BeknurAccount = "0x9Fd598035Ec0DD0909c054E855272b56F2BeC5C8"

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

func NewKeystoreAccount(dataDir, password string) (common.Address, error) {
	ks := keystore.NewKeyStore(GetKeystoreDirPath(dataDir), keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		return common.Address{}, err
	}

	return acc.Address, nil
}

func SignTxWithKeystoreAccount(tx database.Tx, acc common.Address, pwd string) {

}

func Sign(msg []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	msgHash := crypto.Keccak256(msg)

	sig, err = crypto.Sign(msgHash, privKey)
	if err != nil {
		return nil, err
	}

	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength)
	}

	return sig, nil
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	msgHash := crypto.Keccak256(msg)

	recoveredPubKey, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		return nil, fmt.Errorf("unable to verify message signature. %s", err.Error())
	}

	return recoveredPubKey, nil
}
