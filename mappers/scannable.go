package mappers

type scannable interface { // both sql.Rows and sql.Row implement scannable; consolidates mapping
    Scan(...interface{}) error
}
