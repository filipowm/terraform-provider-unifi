# Configure guest access settings for your UniFi network
# This example demonstrates a comprehensive guest portal setup with various authentication options

resource "unifi_portal_file" "logo" {
    file_path = "logo.png"
}

resource "unifi_setting_guest_access" "guest_portal" {
  # Basic configuration
  auth            = "hotspot"  # Authentication type: none, hotspot, custom, or external
  portal_enabled  = true      # Enable the guest portal
  portal_use_hostname = true  # Use hostname for the portal
  portal_hostname = "guest.example.com"  # Portal hostname
  template_engine = "angular" # Portal template engine (angular or jsp)
  
  # Expiration settings for guest access
  expire        = 1440       # Minutes until expiration
  expire_number = 1          # Number of time units
  expire_unit   = 1440       # Time unit in minutes
  
  # Enable external captive portal detection
  ec_enabled    = true
  
  # Password protection for guest access
  password = "guest-access-password"
  
  # Google authentication
  google {
    client_id     = "your-google-client-id"
    client_secret = "your-google-client-secret"
    domain        = "example.com"  # Optional: limit sign-ins to a specific domain
    scope_email   = true         # Request email addresses during sign-in
  }
  
  # Payment option (PayPal)
  payment_gateway = "paypal"
  paypal {
    username    = "business@example.com"
    password    = "paypal-api-password"
    signature   = "paypal-api-signature"
    use_sandbox = true  # Set to false for production
  }
  
  # Redirecting guests after authentication
  redirect {
    url      = "https://example.com/welcome"
    use_https = true
    to_https  = true
  }
  
  # Restricted DNS for guests   
  restricted_dns_servers = [
    "1.1.1.1",
    "8.8.8.8"
  ]
  
  # Portal customization options
  portal_customization {
    customized = true
    
    # Portal appearance
    title = "Welcome to Our Guest Network"
    welcome_text = "Thanks for visiting our location. Please enjoy our complimentary WiFi."
    welcome_text_enabled = true
    welcome_text_position = "top"
    
    # Color scheme
    bg_color = "#f5f5f5"
    text_color = "#333333"
    link_color = "#0078d4"
    
    # Authentication dialog box
    box_color = "#ffffff"
    box_text_color = "#333333"
    box_link_color = "#0078d4"
    box_opacity = 90
    box_radius = 5

    # Logo
    logo_file_id = unifi_portal_file.logo.id
    
    # Button styling
    button_color = "#0078d4"
    button_text_color = "#ffffff"
    button_text = "Connect"
    
    # Legal information / Terms of Service
    tos_enabled = true
    tos = "By using this service, you agree to our terms and conditions. Unauthorized use is prohibited."
    
    # Languages supported
    languages = ["PL"]
  }
}
