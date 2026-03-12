package repository

type Repository interface {
	GetData() string
}

type repositoryImpl struct {
	// Add db connection here
}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) GetData() string {
	// Query database
	return "Data from Database"
}
