# import from provider configured site
terraform import unifi_network.mynetwork 5dc28e5e9106d105bdc87217

# import from another site
terraform import unifi_network.mynetwork zone:5dc28e5e9106d105bdc87217
