# import from provider configured site
terraform import unifi_firewall_zone.myzone 5dc28e5e9106d105bdc87217

# import from another site
terraform import  unifi_firewall_zone.myzone another-site:5dc28e5e9106d105bdc87217
