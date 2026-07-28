# Terraform module for S3 Image Processor Lambda Function
# Triggers specifically on S3 ObjectCreated under raw/ prefix

resource "aws_lambda_function" "image_processor" {
  function_name = var.function_name
  role          = var.lambda_role_arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"
  filename      = var.lambda_zip_path
  memory_size   = 512
  timeout       = 30

  environment {
    variables = {
      CATALOG_SERVICE_URL          = var.catalog_service_url
      LAMBDA_WEBHOOK_SHARED_SECRET = var.lambda_webhook_shared_secret
      CLOUDFRONT_DOMAIN            = var.cloudfront_domain
    }
  }
}

resource "aws_lambda_permission" "allow_s3_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.image_processor.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = var.s3_bucket_arn
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = var.s3_bucket_id

  lambda_function {
    lambda_function_arn = aws_lambda_function.image_processor.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "raw/"
  }

  depends_on = [aws_lambda_permission.allow_s3_bucket]
}
