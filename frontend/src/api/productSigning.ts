import Axios from './https'
import type { ApiResponse } from './customer'
import type { ProductOffering } from './productOffering'

/** 产品公钥响应 */
export interface ProductOfferingPublicKey {
  product_id: string
  product_code: string
  /** PEM 格式公钥；历史产品项可能为空字符串 */
  public_key: string
}

/**
 * 获取产品项公钥
 *
 * 客户端内嵌该公钥以验证许可证签名。私钥不在任何接口暴露。
 */
export function getProductOfferingPublicKey(
  id: string
): Promise<ApiResponse<ProductOfferingPublicKey>> {
  return Axios.get(`/api/v1/product-offerings/${id}/public-key`)
}

export type { ProductOffering }
