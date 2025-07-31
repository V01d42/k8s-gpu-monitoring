export const descendingComparator = <T>(a: T, b: T, orderBy: keyof T) => {
  const aValue = a[orderBy];
  const bValue = b[orderBy];
  if (typeof aValue === "number" && typeof bValue === "number") {
    return bValue - aValue;
  }
  if (typeof aValue === "string" && typeof bValue === "string") {
    return bValue.localeCompare(aValue);
  }
  return 0;
};

export const getComparator = <T>(order: "asc" | "desc", orderBy: keyof T) =>
  order === "desc"
    ? (a: T, b: T) => descendingComparator(a, b, orderBy)
    : (a: T, b: T) => -descendingComparator(a, b, orderBy);
