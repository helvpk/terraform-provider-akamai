provider "akamai" {
  edgerc        = "../../test/edgerc"
  cache_enabled = false
}

data "akamai_botman_custom_bot_category" "test" {
  config_id = 43253
}