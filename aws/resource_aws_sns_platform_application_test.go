package aws

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

/**
 Before running this test, at least one of these ENV variables combinations must be set:

 GCM_API_KEY - Google Cloud Messaging API Key

 APNS_SANDBOX_CREDENTIAL_PATH - Apple Push Notification Sandbox Private Key file location
 APNS_SANDBOX_PRINCIPAL_PATH - Apple Push Notification Sandbox Certificate file location
**/

type testAccAwsSnsPlatformApplicationPlatform struct {
	Name           string
	Credential     string
	CredentialHash string
	Principal      string
	PrincipalHash  string
}

func testAccAwsSnsPlatformApplicationPlatformFromEnv(t *testing.T) []*testAccAwsSnsPlatformApplicationPlatform {
	platforms := make([]*testAccAwsSnsPlatformApplicationPlatform, 0, 2)

	if os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH") != "" {
		if os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH") == "" {
			t.Fatalf("APNS_SANDBOX_CREDENTIAL_PATH set but missing APNS_SANDBOX_PRINCIPAL_PATH")
		}
		credentialHash, err := testAccHashSumPath(os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH"))
		if err != nil {
			t.Fatal(err)
		}
		principalHash, err := testAccHashSumPath(os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH"))
		if err != nil {
			t.Fatal(err)
		}

		platform := &testAccAwsSnsPlatformApplicationPlatform{
			Name:           "APNS_SANDBOX",
			Credential:     fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH")),
			CredentialHash: credentialHash,
			Principal:      fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH")),
			PrincipalHash:  principalHash,
		}
		platforms = append(platforms, platform)
	}

	if os.Getenv("GCM_API_KEY") != "" {
		platform := &testAccAwsSnsPlatformApplicationPlatform{
			Name:           "GCM",
			Credential:     os.Getenv("GCM_API_KEY"),
			CredentialHash: hashSum(os.Getenv("GCM_API_KEY")),
		}
		platforms = append(platforms, platform)
	}

	if len(platforms) == 0 {
		t.Skipf("no SNS Platform Application environment variables found")
	}
	return platforms
}

func testAccHashSumPath(path string) (string, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hashSum(string(data)), nil
}

func TestDecodeResourceAwsSnsPlatformApplicationID(t *testing.T) {

	var testCases = []struct {
		Input            string
		ExpectedArn      string
		ExpectedName     string
		ExpectedPlatform string
		ErrCount         int
	}{
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName",
			ExpectedArn:      "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName",
			ExpectedName:     "myAppName",
			ExpectedPlatform: "APNS_SANDBOX",
			ErrCount:         0,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName/extra",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:endpoint/APNS_SANDBOX/myAppName/someID",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:APNS_SANDBOX/myAppName",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "myAppName",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
	}

	for _, tc := range testCases {
		arn, name, platform, err := decodeResourceAwsSnsPlatformApplicationID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if arn != tc.ExpectedArn {
			t.Fatalf("expected %q to return arn: %s", tc.Input, arn)
		}
		if name != tc.ExpectedName {
			t.Fatalf("expected %q to return name: %s", tc.Input, name)
		}
		if platform != tc.ExpectedPlatform {
			t.Fatalf("expected %q to return platform: %s", tc.Input, platform)
		}
	}
}

func TestAccAwsSnsPlatformApplication_basic(t *testing.T) {
	platforms := testAccAwsSnsPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	for _, platform := range platforms {
		name := fmt.Sprintf("tf-acc-%d", acctest.RandInt())
		platformPrincipalCheck := resource.TestCheckNoResourceAttr(resourceName, "platform_principal")
		if platform.Principal != "" {
			platformPrincipalCheck = resource.TestCheckResourceAttr(resourceName, "platform_principal", platform.PrincipalHash)
		}

		t.Run(platform.Name, func(*testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccCheckAWSSNSPlatformApplicationDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccAwsSnsPlatformApplicationConfig_basic(name, &testAccAwsSnsPlatformApplicationPlatform{
							Name:       "APNS",
							Credential: "NOTEMPTY",
							Principal:  "",
						}),
						ExpectError: regexp.MustCompile(`platform_principal is required when platform =`),
					},
					{
						Config: testAccAwsSnsPlatformApplicationConfig_basic(name, &testAccAwsSnsPlatformApplicationPlatform{
							Name:       "APNS_SANDBOX",
							Credential: "NOTEMPTY",
							Principal:  "",
						}),
						ExpectError: regexp.MustCompile(`platform_principal is required when platform =`),
					},
					{
						Config: testAccAwsSnsPlatformApplicationConfig_basic(name, platform),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckAwsSnsPlatformApplicationExists(resourceName),
							resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:sns:[^:]+:[^:]+:app/%s/%s$", platform.Name, name))),
							resource.TestCheckResourceAttr(resourceName, "name", name),
							resource.TestCheckResourceAttr(resourceName, "platform", platform.Name),
							resource.TestCheckResourceAttr(resourceName, "platform_credential", platform.CredentialHash),
							platformPrincipalCheck,
						),
					},
					{
						ResourceName:            resourceName,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
					},
				},
			})
		})
	}
}

func TestAccAwsSnsPlatformApplication_basicAttributes(t *testing.T) {
	platforms := testAccAwsSnsPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []struct {
		AttributeKey         string
		AttributeValue       string
		AttributeValueUpdate string
	}{
		{
			AttributeKey:         "success_feedback_sample_rate",
			AttributeValue:       "100",
			AttributeValueUpdate: "99",
		},
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc.AttributeKey), func(*testing.T) {
					name := fmt.Sprintf("tf-acc-%d", acctest.RandInt())

					resource.Test(t, resource.TestCase{
						PreCheck:     func() { testAccPreCheck(t) },
						Providers:    testAccProviders,
						CheckDestroy: testAccCheckAWSSNSPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccAwsSnsPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValue),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestCheckResourceAttr(resourceName, tc.AttributeKey, tc.AttributeValue),
								),
							},
							{
								Config: testAccAwsSnsPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValueUpdate),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestCheckResourceAttr(resourceName, tc.AttributeKey, tc.AttributeValueUpdate),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func TestAccAwsSnsPlatformApplication_iamRoleAttributes(t *testing.T) {
	platforms := testAccAwsSnsPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []string{
		"failure_feedback_role_arn",
		"success_feedback_role_arn",
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc), func(*testing.T) {
					iamRoleName1 := fmt.Sprintf("tf-acc-%d", acctest.RandInt())
					iamRoleName2 := fmt.Sprintf("tf-acc-%d", acctest.RandInt())
					name := fmt.Sprintf("tf-acc-%d", acctest.RandInt())

					resource.Test(t, resource.TestCase{
						PreCheck:     func() { testAccPreCheck(t) },
						Providers:    testAccProviders,
						CheckDestroy: testAccCheckAWSSNSPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccAwsSnsPlatformApplicationConfig_iamRoleAttribute(name, platform, tc, iamRoleName1),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s$", iamRoleName1))),
								),
							},
							{
								Config: testAccAwsSnsPlatformApplicationConfig_iamRoleAttribute(name, platform, tc, iamRoleName2),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s$", iamRoleName2))),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func TestAccAwsSnsPlatformApplication_snsTopicAttributes(t *testing.T) {
	platforms := testAccAwsSnsPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []string{
		"event_delivery_failure_topic_arn",
		"event_endpoint_created_topic_arn",
		"event_endpoint_deleted_topic_arn",
		"event_endpoint_updated_topic_arn",
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc), func(*testing.T) {
					snsTopicName1 := fmt.Sprintf("tf-acc-%d", acctest.RandInt())
					snsTopicName2 := fmt.Sprintf("tf-acc-%d", acctest.RandInt())
					name := fmt.Sprintf("tf-acc-%d", acctest.RandInt())

					resource.Test(t, resource.TestCase{
						PreCheck:     func() { testAccPreCheck(t) },
						Providers:    testAccProviders,
						CheckDestroy: testAccCheckAWSSNSPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccAwsSnsPlatformApplicationConfig_snsTopicAttribute(name, platform, tc, snsTopicName1),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:sns:[^:]+:[^:]+:%s$", snsTopicName1))),
								),
							},
							{
								Config: testAccAwsSnsPlatformApplicationConfig_snsTopicAttribute(name, platform, tc, snsTopicName2),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckAwsSnsPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:sns:[^:]+:[^:]+:%s$", snsTopicName2))),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func testAccCheckAwsSnsPlatformApplicationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("missing ID: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).snsconn

		input := &sns.GetPlatformApplicationAttributesInput{
			PlatformApplicationArn: aws.String(rs.Primary.ID),
		}

		log.Printf("[DEBUG] Reading SNS Platform Application attributes: %s", input)
		_, err := conn.GetPlatformApplicationAttributes(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSSNSPlatformApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).snsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_platform_application" {
			continue
		}

		input := &sns.GetPlatformApplicationAttributesInput{
			PlatformApplicationArn: aws.String(rs.Primary.ID),
		}

		log.Printf("[DEBUG] Reading SNS Platform Application attributes: %s", input)
		_, err := conn.GetPlatformApplicationAttributes(input)
		if err != nil {
			if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccAwsSnsPlatformApplicationConfig_basic(name string, platform *testAccAwsSnsPlatformApplicationPlatform) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = "%s"
}
`, name, platform.Name, platform.Credential)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = "%s"
  platform_principal  = "%s"
}
`, name, platform.Name, platform.Credential, platform.Principal)
}

func testAccAwsSnsPlatformApplicationConfig_basicAttribute(name string, platform *testAccAwsSnsPlatformApplicationPlatform, attributeKey, attributeValue string) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = "%s"
  %s                  = "%s"
}
`, name, platform.Name, platform.Credential, attributeKey, attributeValue)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = "%s"
  platform_principal  = "%s"
  %s                  = "%s"
}
`, name, platform.Name, platform.Credential, platform.Principal, attributeKey, attributeValue)
}

func testAccAwsSnsPlatformApplicationConfig_iamRoleAttribute(name string, platform *testAccAwsSnsPlatformApplicationPlatform, attributeKey, iamRoleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {"Service": "sns.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }
}
EOF

  name = "%s"
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
  role       = "${aws_iam_role.test.id}"
}

%s
`, iamRoleName, testAccAwsSnsPlatformApplicationConfig_basicAttribute(name, platform, attributeKey, "${aws_iam_role.test.arn}"))
}

func testAccAwsSnsPlatformApplicationConfig_snsTopicAttribute(name string, platform *testAccAwsSnsPlatformApplicationPlatform, attributeKey, snsTopicName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "%s"
}

%s
`, snsTopicName, testAccAwsSnsPlatformApplicationConfig_basicAttribute(name, platform, attributeKey, "${aws_sns_topic.test.arn}"))
}
