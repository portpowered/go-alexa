package main

type staticError string

func (err staticError) Error() string { return string(err) }

const errNoGraphQLOperations staticError = "genqlient generated no GraphQL operations"
