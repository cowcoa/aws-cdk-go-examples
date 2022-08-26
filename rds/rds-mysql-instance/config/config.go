package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "RdsMysqlCluster"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, set whether to use rds proxy by 'cdk.json/context/enableProxy'.
// The valid value are: true | false
func EnableProxy(scope constructs.Construct) bool {
	enableProxy := false

	ctxValue := scope.Node().TryGetContext(jsii.String("enableProxy"))
	if v, ok := ctxValue.(bool); ok {
		enableProxy = v
	}

	return enableProxy
}

// DO NOT modify this function, set whether to use rds replica by 'cdk.json/context/enableReplica'.
// The valid value are: true | false
func EnableReplica(scope constructs.Construct) bool {
	enableReplica := false

	ctxValue := scope.Node().TryGetContext(jsii.String("enableReplica"))
	if v, ok := ctxValue.(bool); ok {
		enableReplica = v
	}

	return enableReplica
}

// DO NOT modify this function, set multi az by 'cdk.json/context/enableMultiAz'.
func EnableMultiAz(scope constructs.Construct) bool {
	enableMultiAz := false

	ctxValue := scope.Node().TryGetContext(jsii.String("enableMultiAz"))
	if v, ok := ctxValue.(bool); ok {
		enableMultiAz = v
	}

	return enableMultiAz
}

// DO NOT modify this function, set mysql version by 'cdk.json/context/mysqlVersion'.
func MySQLVersion(scope constructs.Construct) float64 {
	mysqlVersion := 5.7

	ctxValue := scope.Node().TryGetContext(jsii.String("mysqlVersion"))
	if v, ok := ctxValue.(float64); ok {
		mysqlVersion = v
	}

	return mysqlVersion
}

// DO NOT modify this function, change event subscription by 'cdk.json/context/eventSubEmail'.
func EventSubEmail(scope constructs.Construct) string {
	eventSubEmail := "aws@amazon.com"

	ctxValue := scope.Node().TryGetContext(jsii.String("eventSubEmail"))
	if v, ok := ctxValue.(string); ok {
		eventSubEmail = v
	}

	return eventSubEmail
}

// DO NOT modify this function, set db master user by 'cdk.json/context/masterUser'.
func MasterUser(scope constructs.Construct) string {
	masterUser := "admin"

	ctxValue := scope.Node().TryGetContext(jsii.String("masterUser"))
	if v, ok := ctxValue.(string); ok {
		masterUser = v
	}

	return masterUser
}

// DO NOT modify this function, set init database by 'cdk.json/context/initDatabase'.
func InitDatabase(scope constructs.Construct) string {
	initDatabase := "mydb"

	ctxValue := scope.Node().TryGetContext(jsii.String("initDatabase"))
	if v, ok := ctxValue.(string); ok {
		initDatabase = v
	}

	return initDatabase
}
