import { EnumMap } from "src/models/enum_map";
export function enumToArray<Enum>(e: Enum): EnumMap<keyof Enum, Enum>[] {
  const collection: EnumMap<keyof Enum, Enum>[] = [];
  let entries = Object.entries(e);
  entries = entries.slice(entries.length / 2);
  for (const key in entries) {
    collection.push({ key: key as keyof Enum, value: e[key] });
  }
  return collection;
}

export function arrayToEnum<E>(a: EnumMap<keyof E, E>[], location: string): E[] {
  const Unkown = require(location) as E
  return a.map((value)=> value.value[value.key] as E)
}