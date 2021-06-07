package cantabular

// coder is an interface that allows you to 
// extract a http status code from an error (or other object)
type coder interface{
	Code() int
}