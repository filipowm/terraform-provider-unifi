resource "unifi_setting_country" "example" {
  # Set the country code using ISO 3166-1 alpha-2 format
  # This example sets the country to United States
  code = "US"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
