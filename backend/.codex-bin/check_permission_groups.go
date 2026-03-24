package main
import (
  "fmt"
  "github.com/gg-ecommerce/backend/internal/config"
  "github.com/gg-ecommerce/backend/internal/pkg/database"
)
func main() {
  cfg, err := config.Load(); if err != nil { panic(err) }
  _, err = database.Init(&cfg.DB); if err != nil { panic(err) }
  defer database.Close()
  var exists int64
  if err := database.DB.Raw("select count(*) from information_schema.tables where table_schema = current_schema() and table_name = 'permission_groups'").Scan(&exists).Error; err != nil { panic(err) }
  fmt.Println("permission_groups_table_exists", exists)
  if exists > 0 {
    var groupCount int64
    var actionCount int64
    if err := database.DB.Raw("select count(*) from permission_groups").Scan(&groupCount).Error; err != nil { panic(err) }
    if err := database.DB.Raw("select count(*) from permission_actions where module_group_id is not null and feature_group_id is not null").Scan(&actionCount).Error; err != nil { panic(err) }
    fmt.Println("permission_groups_count", groupCount)
    fmt.Println("permission_actions_with_groups", actionCount)
  }
}
