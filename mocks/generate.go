//go:generate ../bin/minimock -i github.com/WithSoull/in-memory-database/internal/database.computeLayer -o ./ -s "_mock.go" -g
//go:generate ../bin/minimock -i github.com/WithSoull/in-memory-database/internal/database.storageLayer -o ./ -s "_mock.go" -g
//go:generate ../bin/minimock -i github.com/WithSoull/in-memory-database/internal/database/compute.Query -o ./ -s "_mock.go" -g
//go:generate ../bin/minimock -i github.com/WithSoull/in-memory-database/internal/database/storage.Engine -o ./ -s "_mock.go" -g
//go:generate ../bin/minimock -i github.com/WithSoull/in-memory-database/internal/database/storage/engine/in_memory.Hashtable -o ./ -s "_mock.go" -g

package mocks
