package xcrypto

import (
	"reflect"

	"github.com/liamylian/x-rsa/golang/xrsa"
)

// DecryptStructWithXRsa ...
func DecryptStructWithXRsa(data interface{}, xr *xrsa.XRsa) error {
	v := reflect.ValueOf(data).Elem()
	t := reflect.TypeOf(data).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.String {
			plainText, err := xr.PrivateDecrypt(v.Field(i).String())
			if err != nil {
				return err
			}
			v.Field(i).SetString(plainText)
		}
	}
	return nil
}

// EncryptStructWithXRsa ..
func EncryptStructWithXRsa(data interface{}, xr *xrsa.XRsa) error {
	v := reflect.ValueOf(data).Elem()
	t := reflect.TypeOf(data).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.String {
			encryptedText, err := xr.PublicEncrypt(v.Field(i).String())
			if err != nil {
				return err
			}
			v.Field(i).SetString(encryptedText)
		}
	}
	return nil
}
