package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceFastlyFlattenDictionary(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Dictionary
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Dictionary{
				{
					ID:        "1234567890",
					Name:      "dictionary-example",
					WriteOnly: true,
				},
			},
			local: []map[string]interface{}{
				{
					"id":         "1234567890",
					"name":       "dictionary-example",
					"write_only": true,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDictionaries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_dictionary(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_dictionary(name, dictName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, name, dictName),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_dictionary(service *gofastly.ServiceDetail, name, dictName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dict, err := conn.GetDictionary(&gofastly.GetDictionaryInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
			Name:    dictName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if dict.Name != dictName {
			return fmt.Errorf("Dictionary logging endpoint name mismatch, expected: %s, got: %#v", dictName, dict.Name)
		}

		return nil
	}
}

func testAccServiceV1Config_dictionary(name, dictName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  dictionary {
	name       = "%s"
	write_only = false
  }

  force_destroy = true
}`, name, domainName, backendName, dictName)
}
