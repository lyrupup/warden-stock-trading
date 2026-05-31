/** 用户（对应 openapi User） */
export type TUser = {
  id: number;
  username: string;
  nickname?: string;
  avatar?: string;
  status?: number;
};
