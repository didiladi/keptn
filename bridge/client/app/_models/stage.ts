import {Service} from "./service";

export class Stage {
  stageName: string;
  services: Service[];

  public getOpenApprovals() {
    return this.services.reduce((openApprovals, service) => [...openApprovals, ...service.getOpenApprovals()], []);
  }

  public getOpenProblems() {
    return this.services.reduce((openProblems, service) => [...openProblems, ...service.getOpenProblems()], []);
  }

  static fromJSON(data: any) {
    return Object.assign(new this, data);
  }
}
