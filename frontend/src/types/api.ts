/** 统一响应体（见 FRONTEND.md §7 / openapi ApiResponse），code===0 为成功 */
export type TApiResponse<T = unknown> = {
  code: number;
  message: string;
  data: T;
};

/** 分页响应体（对应 openapi PageResult） */
export type TPageResult<T> = {
  list: T[];
  total: number;
  page: number;
  size: number;
};

/** 分页查询参数 */
export type TPageParams = {
  page?: number;
  size?: number;
};
