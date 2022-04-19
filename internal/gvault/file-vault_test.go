package vault

import (
	"fmt"
	"testing"
)

func TestVault_Init(t *testing.T) {
	t.Setenv("CDT_VAULT_SECRET", "very_secret")

	_, err := CreateVault("unit-test")
	if err != nil {
		t.Error(err)
	}
}

func TestVault_WriteRead(t *testing.T) {
	t.Setenv("CDT_VAULT_SECRET", "very_secret")

	fv, err := CreateVault("unit-test")
	if err != nil {
		t.Error(err)
	}

	err = fv.Write("mykey", "myvalue")
	if err != nil {
		t.Error(err)
	}

	val, err := fv.Read("mykey")
	if err != nil {
		t.Error(err)
	}

	if val != "myvalue" {
		t.Errorf("wrong value")
	}
}

func TestVault_MultiWriteRead(t *testing.T) {
	t.Setenv("CDT_VAULT_SECRET", "very_secret")

	fv, err := CreateVault("unit-test")
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = fv.Write(fmt.Sprintf("mykey-%d", i), fmt.Sprintf("myvalue-%d", i))
		if err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 1000; i++ {
		val, err := fv.Read(fmt.Sprintf("mykey-%d", i))
		if err != nil {
			t.Error(err)
		}

		if val != fmt.Sprintf("myvalue-%d", i) {
			t.Errorf("wrong value")
		}
	}
}
