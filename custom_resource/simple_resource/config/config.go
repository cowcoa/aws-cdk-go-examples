package config

const (
	StackName = "CdkGolangExample-SimpleCustomResource"
	// Lambda function config
	RoleName    = "CRLambdaRole"
	FuncionName = "CRLambdaFunction"
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
