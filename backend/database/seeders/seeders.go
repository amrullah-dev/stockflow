package seeders

import "github.com/goravel/framework/contracts/database/seeder"

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&DevelopmentSeeder{},
	}
}
