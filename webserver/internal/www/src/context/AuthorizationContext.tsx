import { type ReactNode, createContext } from "react";
import { gql, useQuery } from "@apollo/client";
// import { UserType } from "../utils/consts";

export type UserType = {
    id: string;
    name: string;
    isactivated?: boolean;
}

export type AuthorizationContextType = {
    me: UserType;
}
export type AuthorizationContextQueryType = {
    data: undefined | AuthorizationContextType;
    isLoading: boolean;
    error: any;
}

const defaultValue = { data: undefined, isLoading: false, error: undefined } as AuthorizationContextQueryType;

export const AuthorizationContext = createContext(defaultValue);

export const AuthorizationContextProvider = ({ children }: { children: ReactNode }) => {

    const GET_USER_INFO = gql`
        query GetMe{
            me {
                id
                name
                isactivated
            }
        }
    `;

    const { loading: isLoading, error, data } = useQuery(GET_USER_INFO);

    return (
        <AuthorizationContext.Provider value={{ data, isLoading, error }}>
            {children}
        </AuthorizationContext.Provider>
    );
};