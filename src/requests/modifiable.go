/**
 * Civil Recovery Service
 *
 * @author    Ashley Kitson
 * @license   Proprietary. See 'Licence.pdf'
 * @copyright Zinc Digital Business Solutions Ltd, UK, 2022
 */

package requests

import "github.com/aws/aws-lambda-go/events"

func ModifyRequest(request events.APIGatewayProxyRequest, ctx string) events.APIGatewayProxyRequest {
	newRequest := request
	newRequest.IsBase64Encoded = false
	//set context id
	headers := newRequest.Headers
	headers["x-context-id"] = ctx
	newRequest.Headers = headers

	return newRequest
}
