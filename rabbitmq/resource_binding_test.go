package rabbitmq

import (
	"fmt"
	"strings"
	"testing"

	"github.com/michaelklishin/rabbit-hole"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBinding_basic(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBindingConfig_basic,
				Check: testAccBindingCheck(
					"rabbitmq_binding.test", &bindingInfo,
				),
			},
		},
	})
}

func TestAccBinding_basic_on_slash_vhost(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBindingConfig_basic_slash_vhost,
				Check: testAccBindingCheck(
					"rabbitmq_binding.slashtest", &bindingInfo,
				),
			},
		},
	})
}

func TestAccBinding_propertiesKey(t *testing.T) {
	var bindingInfo rabbithole.BindingInfo
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccBindingCheckDestroy(bindingInfo),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBindingConfig_propertiesKey,
				Check: testAccBindingCheck(
					"rabbitmq_binding.test", &bindingInfo,
				),
			},
		},
	})
}

func testAccBindingCheck(rn string, bindingInfo *rabbithole.BindingInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("binding id not set")
		}

		rmqc := testAccProvider.Meta().(*rabbithole.Client)
		bindingParts := strings.Split(rs.Primary.ID, "/")

		index := 0
		vhost := bindingParts[0]

		// vhost is /, so we need to skip the first item, and replace empty vhost
		if strings.HasPrefix(rs.Primary.ID, "//") {
			index = 1
			vhost = "/"
			if len(bindingParts) < 6 {
				return fmt.Errorf("Unable to determine binding ID")
			}
		}

		bindings, err := rmqc.ListBindingsIn(vhost)
		if err != nil {
			return fmt.Errorf("Error retrieving exchange: %s", err)
		}

		for _, binding := range bindings {
			if binding.Source == bindingParts[index+1] && binding.Destination == bindingParts[index+2] && binding.DestinationType == bindingParts[index+3] && binding.PropertiesKey == bindingParts[index+4] {
				bindingInfo = &binding
				return nil
			}
		}

		return fmt.Errorf("Unable to find binding %s", rn)
	}
}

func testAccBindingCheckDestroy(bindingInfo rabbithole.BindingInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rmqc := testAccProvider.Meta().(*rabbithole.Client)

		bindings, err := rmqc.ListBindingsIn(bindingInfo.Vhost)
		if err != nil {
			return fmt.Errorf("Error retrieving exchange: %s", err)
		}

		for _, binding := range bindings {
			if binding.Source == bindingInfo.Source && binding.Destination == bindingInfo.Destination && binding.DestinationType == bindingInfo.DestinationType && binding.PropertiesKey == bindingInfo.PropertiesKey {
				return fmt.Errorf("Binding still exists")
			}
		}

		return nil
	}
}

const testAccBindingConfig_basic = `
resource "rabbitmq_vhost" "test" {
    name = "test"
}

resource "rabbitmq_permissions" "guest" {
    user = "guest"
    vhost = "${rabbitmq_vhost.test.name}"
    permissions {
        configure = ".*"
        write = ".*"
        read = ".*"
    }
}

resource "rabbitmq_exchange" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "#"
}`

const testAccBindingConfig_basic_slash_vhost = `
resource "rabbitmq_vhost" "slashtest" {
    name = "/"
}

resource "rabbitmq_permissions" "guest" {
    user = "guest"
    vhost = "${rabbitmq_vhost.slashtest.name}"
    permissions {
        configure = ".*"
        write = ".*"
        read = ".*"
    }
}

resource "rabbitmq_exchange" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "fanout"
        durable = false
        auto_delete = true
    }
}

resource "rabbitmq_queue" "test" {
    name = "test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "slashtest" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.slashtest.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "#"
}`

const testAccBindingConfig_propertiesKey = `
resource "rabbitmq_vhost" "test" {
    name = "test"
}

resource "rabbitmq_permissions" "guest" {
    user = "guest"
    vhost = "${rabbitmq_vhost.test.name}"
    permissions {
        configure = ".*"
        write = ".*"
        read = ".*"
    }
}

resource "rabbitmq_exchange" "test" {
    name = "Test"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        type = "topic"
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_queue" "test" {
    name = "Test.Queue"
    vhost = "${rabbitmq_permissions.guest.vhost}"
    settings {
        durable = true
        auto_delete = false
    }
}

resource "rabbitmq_binding" "test" {
    source = "${rabbitmq_exchange.test.name}"
    vhost = "${rabbitmq_vhost.test.name}"
    destination = "${rabbitmq_queue.test.name}"
    destination_type = "queue"
    routing_key = "ANYTHING.#"
    arguments = {
      key1 = "value1"
      key2 = "value2"
      key3 = "value3"
    }
}
`
