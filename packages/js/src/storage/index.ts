export * from "./adapter.js";
export { createLocalFS } from "./local/local.js";
export { createS3FS, DRIVER_S3, type S3Config, type S3FS, type S3API, type S3Presigner } from "./s3/s3.js";
