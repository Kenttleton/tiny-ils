import { Injectable } from "@nestjs/common";
import { v4 } from "uuid";

// This should be a real class/interface representing a user entity
export type User = any;

@Injectable()
export class UsersService {
  private readonly users = [
    {
      identity: v4(),
      username: "johnsmith",
      password: "changeme",
    },
    {
      identity: v4(),
      username: "mariagonzalez",
      password: "guess",
    },
  ];

  async findOne(username: string): Promise<User | undefined> {
    return this.users.find((user) => user.username === username);
  }
}
