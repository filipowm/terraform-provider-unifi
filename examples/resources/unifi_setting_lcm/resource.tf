resource "unifi_setting_lcd_monitor" "example" {
  # Enable LCD monitor functionality
  enabled = true
  
  # Set the brightness level (0-100)
  brightness = 75
  
  # Set the idle timeout in seconds before the display dims
  idle_timeout = 300
  
  # Enable synchronization of settings across all devices
  sync = true
  
  # Enable touch events on the LCD screen
  touch_event = true
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
