package storage

type Storage interface {
	Get(key string) (string, bool)
	Set(key, value string, ttlSeconds *int64) error
	Delete(key string) (bool, error)
	List(limit int) (map[string]string, error)
}
