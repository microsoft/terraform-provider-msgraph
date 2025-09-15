terraform {
  required_providers {
    msgraph = {
      source = "Microsoft/msgraph"
    }
  }
}

provider "msgraph" {}

# Example 1: Send a welcome email to a user
resource "msgraph_resource_action" "send_welcome_email" {
  resource_url = "users/john@example.com"
  action       = "sendMail"
  method       = "POST"

  body = {
    message = {
      subject = "Welcome to the organization!"
      body = {
        contentType = "HTML"
        content     = "<h1>Welcome!</h1><p>We're excited to have you join our team.</p>"
      }
      toRecipients = [
        {
          emailAddress = {
            address = "john@example.com"
            name    = "John Doe"
          }
        }
      ]
    }
    saveToSentItems = true
  }
}

# Example 2: Reset a user's password
resource "msgraph_resource_action" "reset_user_password" {
  resource_url = "users/jane@example.com"
  action       = "changePassword"
  method       = "POST"

  body = {
    currentPassword = "OldPassword123!"
    newPassword     = "NewSecurePassword456!"
  }
}

# Example 3: Send a custom notification with specific headers
resource "msgraph_resource_action" "send_notification" {
  resource_url = "communications/calls"
  action       = "logTeleconferenceDeviceQuality"
  method       = "POST"

  headers = {
    "ConsistencyLevel" = "eventual"
    "X-AnchorMailbox"  = "john@example.com"
  }

  body = {
    quality = {
      timestamp = "2024-01-01T00:00:00Z"
      data      = "quality-data"
    }
  }
}

resource "msgraph_resource_action" "custom_user_action" {
  resource_url = "users/john@example.com"
  action       = "customAction"
  method       = "POST"

  query_parameters = {
    "$select" = ["id", "displayName"]
    "force"   = ["true"]
  }

  headers = {
    "X-Custom-Header"  = "custom-value"
    "X-Request-Source" = "terraform"
    "Content-Language" = "en-US"
  }

  body = {
    actionType = "sync"
    parameters = {
      syncAll = true
    }
  }

  response_export_values = {
    result_id   = "id"
    status      = "status"
    all_details = "@"
  }

  retry = {
    error_message_regex = [
      ".*throttled.*",
      ".*rate limit.*"
    ]
  }
}

# Output the results
output "welcome_email_sent" {
  value = msgraph_resource_action.send_welcome_email.output
}

output "notification_sent" {
  value = msgraph_resource_action.send_notification.output
}

output "custom_action_result" {
  value = msgraph_resource_action.custom_user_action.output
}
