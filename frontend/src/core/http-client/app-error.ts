/** 业务错误：当统一响应体 code !== 0 时抛出 */
export class AppError extends Error {
  constructor(
    public code: number,
    message: string,
  ) {
    super(message);
    this.name = "AppError";
  }
}
