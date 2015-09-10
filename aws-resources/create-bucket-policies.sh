set -u

. common.rc

check


USER=$APPLICATION_NAME

ACCOUNT_ID=$(aws --profile $AWS_PROFILE iam get-user | jq -r ".User.UserId")

CONSOLE_LOGS_BUCKET_POLICY=$(cat <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": "s3:ListBucket",
			"Resource": "arn:aws:s3:::${APPLICATION_NAME}-console-logs"
		},
		{
			"Effect": "Allow",
			"Action": [
				"s3:DeleteObject",
				"s3:GetObject",
				"s3:PutObject"
			],
			"Resource": "arn:aws:s3:::${APPLICATION_NAME}-console-logs/*"
		}
	]
}
EOF
)

BUILD_ARTIFACTS_BUCKET_POLICY=$(cat <<EOF
{
        "Version": "2012-10-17",
        "Statement": [
                {
                        "Effect": "Allow",
                        "Action": "s3:ListBucket",
                        "Resource": "arn:aws:s3:::${APPLICATION_NAME}-build-artifacts"
                },
                {
                        "Effect": "Allow",
                        "Action": [
                                "s3:DeleteObject",
                                "s3:GetObject",
                                "s3:PutObject"
                        ],
                        "Resource": "arn:aws:s3:::${APPLICATION_NAME}-build-artifacts/*"
                }
        ]
}
EOF
)


echo "===Creating S3 bucket policies"

CONSOLE_LOGS=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-s3-console-logs --policy-document "$CONSOLE_LOGS_BUCKET_POLICY" --description "Give r/w to $USER user on S3 console logs bucket")

BUILD_ARTIFACTS=$(aws --profile $AWS_PROFILE iam create-policy --policy-name ${APPLICATION_NAME}-s3-build-artifacts --policy-document "$BUILD_ARTIFACTS_BUCKET_POLICY" --description "Give r/w to $USER user on S3 build artifacts bucket")

for i in "$CONSOLE_LOGS" "$BUILD_ARTIFACTS" ; do 
	aws --profile $AWS_PROFILE iam attach-user-policy --user-name $USER --policy-arn "$(echo "$i" | jq -r ".Policy.Arn")"
done

