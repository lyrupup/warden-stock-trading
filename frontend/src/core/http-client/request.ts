import type { Options } from "ky";
import type { TApiResponse } from "@/types";
import { httpClient } from "./http-client";

/**
 * 统一响应解包：把 ky 返回的 {code,message,data} 取出 data。
 * 业务 code !== 0 已在 http-client 的 afterResponse 中抛出 AppError，
 * 因此这里 resolve 时一定是成功响应。
 */
async function unwrap<T>(promise: Promise<TApiResponse<T>>): Promise<T> {
  const body = await promise;
  return body.data;
}

/** 语义化请求方法，返回已解包的 data（业务层优先使用） */
export const http = {
  get: <T>(url: string, options?: Options): Promise<T> =>
    unwrap<T>(httpClient.get(url, options).json<TApiResponse<T>>()),
  post: <T>(url: string, options?: Options): Promise<T> =>
    unwrap<T>(httpClient.post(url, options).json<TApiResponse<T>>()),
  put: <T>(url: string, options?: Options): Promise<T> =>
    unwrap<T>(httpClient.put(url, options).json<TApiResponse<T>>()),
  patch: <T>(url: string, options?: Options): Promise<T> =>
    unwrap<T>(httpClient.patch(url, options).json<TApiResponse<T>>()),
  delete: <T>(url: string, options?: Options): Promise<T> =>
    unwrap<T>(httpClient.delete(url, options).json<TApiResponse<T>>()),
};
