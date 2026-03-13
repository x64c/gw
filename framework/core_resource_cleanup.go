package framework

import (
	"log"
)

func (c *Core) ResourceCleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")
	// Clean up DB clients ----
	// ToDo: factor out this
	for name, kvdbClient := range c.KVDBClients {
		log.Printf("[INFO] Closing KVDB client %q", name)
		if err := kvdbClient.Close(); err != nil {
			log.Printf("[ERROR] Failed to close KVDB client %q: %v", name, err)
		}
	}
	for name, sqlDBClient := range c.SQLDBClients {
		dbType := sqlDBClient.Conf().Type
		log.Printf("[INFO][%s] Closing %q SQL DB client", dbType, name)
		err := sqlDBClient.Close()
		if err != nil {
			log.Printf("[ERROR][%s] Failed to close %q SQL DB client", dbType, name)
		} else {
			log.Printf("[INFO][%s] %q SQL DB client closed", dbType, name)
		}
	}
	//----
	log.Println("[INFO] App Resource Cleanup Complete")
}
