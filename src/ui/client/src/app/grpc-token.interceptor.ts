export class GrpcTokenInterceptor {
  constructor(private _token: string) {}

  // TODO: The @improbable-en/grpc-web impl does not support
  // client interceptors. So we mimick an interceptor by returning metadata
  // that contains the grpc auth token.
  get intercept() {
    return { Authorization: `Bearer ${this._token}` };
  }
}
