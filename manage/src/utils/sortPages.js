//ページを名前順に並べ、同じ名前どうしは作成日の新しい順にする（Array.prototype.sort 用）
export const byNameThenNewest = (a, b) => (
  (a.name || '').localeCompare(b.name || '', 'ja')
    || new Date(b.createdAt || 0) - new Date(a.createdAt || 0)
)
