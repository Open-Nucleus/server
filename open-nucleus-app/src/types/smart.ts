/** A SMART on FHIR client application registration. */
export interface SmartClient {
  id: string;
  name: string;
  client_uri?: string;
  redirect_uri?: string;
  scope?: string;
}
