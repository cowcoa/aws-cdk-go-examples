package config

const (
	StackName = "CdkGolangExample-ApiGtwLambdaDdb"
	// Lambda function config
	RoleName    = "CRLambdaRole"
	FuncionName = "ApigLambdaFunction"
	MemorySize  = 128
	MaxDuration = 60
	CodePath    = "function/."
	Handler     = "main"
	// Provider function config
	ProviderName = "CRProvider"
	// Custom resource config
	ResourceName = "SSMParamCustomRes"
	ResourceType = "Custom::SSMParamCustomRes"
	// SSM Config
	PhysicalIdKey    = "PhysicalResourceId"
	SsmParamNameKey  = "SSMParamName"
	SsmParamValueKey = "SSMParamValue"
)
