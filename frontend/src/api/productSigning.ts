import Axios from './https'
import type { ApiResponse } from './types'
import type { ProductOffering } from './productOffering'

/** 签名密钥链条目 */
export interface SigningKeyChainEntry {
  /** 本条公钥 PEM */
  public_key: string
  /** 前任私钥对本公钥的签名（base64）；根条目为空 */
  signature?: string
  /** 退役时间；当前签发密钥为空 */
  retired_at?: string
  /** 轮换原因 */
  reason?: string
}

/** 产品公钥响应 */
export interface ProductOfferingPublicKey {
  product_id: string
  product_code: string
  /** PEM 格式公钥；历史产品项可能为空字符串 */
  public_key: string
  /** 全部有效公钥（当前 + 轮换留档），顺序：当前在前 */
  public_keys?: string[]
  /** 完整签名密钥链（含每级衔接签名） */
  key_chain?: SigningKeyChainEntry[]
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

/**
 * 轮换产品项签名密钥
 *
 * 生成新密钥对作为当前签发密钥，旧公钥留档并由旧私钥为新公钥背书。
 * 客户端凭内置根公钥逐级验证密钥链即可纳入新公钥，无需重新发版。
 */
export function rotateProductOfferingSigningKey(
  id: string,
  reason?: string
): Promise<ApiResponse<ProductOfferingPublicKey>> {
  return Axios.post(`/api/v1/product-offerings/${id}/rotate-signing-key`, {
    reason: reason ?? ''
  })
}

export type { ProductOffering }
