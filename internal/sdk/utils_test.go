package sdk

import (
	"github.com/easy-model-fusion/client/internal/config"
	"github.com/easy-model-fusion/client/internal/utils"
	"github.com/easy-model-fusion/client/test"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestCheckForUpdates(t *testing.T) {
	dname := test.CreateFullTestSuite(t)
	defer os.RemoveAll(dname)

	err := config.GetViperConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	viper.Set("sdk-tag", "")
	test.AssertEqual(t, checkForUpdates(), false, "Should return false if no tag is set")

	viper.Set("sdk-tag", "v0.0.1")
	test.AssertEqual(t, checkForUpdates(), true, "Should return true if tag is set and there is an update")

	tag, err := utils.GetLatestTag("sdk")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	viper.Set("sdk-tag", tag)
	test.AssertEqual(t, checkForUpdates(), false, "Should return false if tag is set and there is no update")
}

func TestCanSendUpdateSuggestion(t *testing.T) {
	dname := test.CreateFullTestSuite(t)
	defer os.RemoveAll(dname)

	err := config.GetViperConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	test.AssertEqual(t, canSendUpdateSuggestion(), true, "Should return true if update-suggested is not set")

	viper.Set("update-suggested", false)
	test.AssertEqual(t, canSendUpdateSuggestion(), true, "Should return true if update-suggested is false")

	viper.Set("update-suggested", true)
	test.AssertEqual(t, canSendUpdateSuggestion(), false, "Should return false if update-suggested is true")
}

func TestResetUpdateSuggestion(t *testing.T) {
	dname := test.CreateFullTestSuite(t)
	defer os.RemoveAll(dname)

	err := config.GetViperConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	viper.Set("update-suggested", true)
	ResetUpdateSuggestion()
	test.AssertEqual(t, viper.GetBool("update-suggested"), false, "Should set update-suggested to false")
}

func TestSetUpdateSuggestion(t *testing.T) {
	dname := test.CreateFullTestSuite(t)
	defer os.RemoveAll(dname)

	err := config.GetViperConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	setUpdateSuggestion(true)
	test.AssertEqual(t, viper.GetBool("update-suggested"), true, "Should set update-suggested to true")

	setUpdateSuggestion(false)
	test.AssertEqual(t, viper.GetBool("update-suggested"), false, "Should set update-suggested to false")
}

func TestSendUpdateSuggestion(t *testing.T) {
	dname := test.CreateFullTestSuite(t)
	defer os.RemoveAll(dname)

	err := config.GetViperConfig()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	viper.Set("update-suggested", false)
	viper.Set("sdk-tag", "")
	SendUpdateSuggestion()
	test.AssertEqual(t, viper.GetBool("update-suggested"), false, "Should not set update-suggested to true if there is no tag")

	viper.Set("update-suggested", false)
	viper.Set("sdk-tag", "v0.0.1")
	SendUpdateSuggestion()
	test.AssertEqual(t, viper.GetBool("update-suggested"), true, "Should set update-suggested to true if there is a tag and update-suggested is false")

	viper.Set("update-suggested", true)
	viper.Set("sdk-tag", "v0.0.1")
	SendUpdateSuggestion()
	test.AssertEqual(t, viper.GetBool("update-suggested"), true, "Should not set update-suggested to true if there is a tag and update-suggested is true")
}