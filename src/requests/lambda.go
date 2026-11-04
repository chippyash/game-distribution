package requests

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

/*
*
Provide mockable interface to grab lambda context
*/
type LambdaContextable interface {
	FromContext(ctx context.Context) (*lambdacontext.LambdaContext, bool)
}
